package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-ubipay-websocket/config"
	"go-ubipay-websocket/models"

	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Database struct {
	Client                *mongo.Client
	UserWalletCollection  *mongo.Collection
	TransactionCollection *mongo.Collection
	TestMode              bool
	User                  *mongo.Collection
	mockWallets           map[primitive.ObjectID]*models.UserWallet
	mockTransactions      []*models.TransactionMovement
}

var DB *Database

func ConnectMongoDB(cfg *config.Config) (*Database, error) {
	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(cfg.MongoDBURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("❌ MongoDB connection failed: %v", err)
		return nil, err
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("❌ MongoDB ping failed: %v", err)
		return nil, err
	}

	db := client.Database(cfg.MongoDBName)

	database := &Database{
		Client:                client,
		UserWalletCollection:  db.Collection("TblUserWallet"),
		TransactionCollection: db.Collection("TblTransactionMovement"),
		User:                  db.Collection("TblUser"),
		TestMode:              false,
	}

	DB = database
	log.Println("✅ MongoDB connected successfully")
	return database, nil
}

// NewTestDatabase creates a mock database for testing without MongoDB
func NewTestDatabase() *Database {
	log.Println("🔧 Using test database (MongoDB not available)")
	return &Database{
		mockWallets:      make(map[primitive.ObjectID]*models.UserWallet),
		mockTransactions: make([]*models.TransactionMovement, 0),
		TestMode:         true,
	}
}

func (db *Database) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Client.Disconnect(ctx)
}

func (db *Database) GetUserWallet(userID primitive.ObjectID) (*models.UserWallet, error) {
	if db.TestMode {
		// Mock implementation for test mode
		if wallet, exists := db.mockWallets[userID]; exists {
			log.Printf("🔍 [TEST] Retrieved wallet for user %s - Balance: %d", userID.Hex(), wallet.Balance)
			return wallet, nil
		}
		// Create a new wallet if it doesn't exist
		return db.CreateUserWallet(userID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wallet models.UserWallet
	filter := bson.M{"UserID": userID, "WalletType": 1, "Enable": true}

	err := db.UserWalletCollection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create a new wallet if it doesn't exist
			return db.CreateUserWallet(userID)
		}
		return nil, err
	}

	return &wallet, nil
}

func (db *Database) GetUserBySessionToken(sessionToken string) (*models.User, error) {
	// Check if User collection is available (test mode or not initialized)
	if db.User == nil {
		return nil, fmt.Errorf("user collection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	filter := bson.M{"SessionToken": sessionToken}

	err := db.User.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Decimal128 → int
func decimal128ToInt(d primitive.Decimal128) int {
	s := d.String() // "3.0"
	s = strings.Split(s, ".")[0]
	i, _ := strconv.Atoi(s)
	return i
}

// int → Decimal128
func intToDecimal128(i int) primitive.Decimal128 {
	d, _ := primitive.ParseDecimal128(strconv.Itoa(i))
	return d
}

func (db *Database) CreateUserWallet(userID primitive.ObjectID) (*models.UserWallet, error) {
	wallet := models.UserWallet{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		WalletType:   1,
		WalletName:   "Point Wallet",
		Balance:      intToDecimal128(0),
		Enable:       true,
		CreateBy:     "System",
		CreateDate:   time.Now(),
		ModifiedBy:   "System",
		ModifiedDate: time.Now(),
	}

	if db.TestMode {
		// Mock implementation for test mode
		db.mockWallets[userID] = &wallet
		log.Printf("➕ [TEST] Created new wallet for user %s", userID.Hex())
		return &wallet, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.UserWalletCollection.InsertOne(ctx, wallet)
	if err != nil {
		return nil, err
	}

	wallet.ID = result.InsertedID.(primitive.ObjectID)
	log.Printf("➕ Created new wallet for user %s", userID.Hex())
	return &wallet, nil
}

func (db *Database) UpdateWalletBalance(userID primitive.ObjectID, amount int) (*models.UserWallet, error) {
	wallet, err := db.GetUserWallet(userID)
	if err != nil {
		return nil, err
	}

	current := decimal128ToInt(wallet.Balance)
	newBalance := current + amount
	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	wallet.Balance = intToDecimal128(newBalance)
	wallet.ModifiedBy = "API"
	wallet.ModifiedDate = time.Now()

	if db.TestMode {
		db.mockWallets[userID] = wallet
		log.Printf("💵 [TEST] Updated wallet balance for user %s: %+d → %+d",
			userID.Hex(), current, newBalance)
		return wallet, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"Balance":      wallet.Balance,
			"ModifiedBy":   wallet.ModifiedBy,
			"ModifiedDate": wallet.ModifiedDate,
		},
	}

	_, err = db.UserWalletCollection.UpdateOne(ctx, bson.M{"_id": wallet.ID}, update)
	if err != nil {
		return nil, err
	}

	log.Printf("💵 Updated wallet balance for user %s: %+d → %+d",
		userID.Hex(), current, newBalance)

	return wallet, nil
}

func (db *Database) CreateTransaction(userID primitive.ObjectID, username string, transactionType, targetType, amount, beforeAmt, afterAmt int) error {
	transaction := models.TransactionMovement{
		ID:              primitive.NewObjectID(),
		UserID:          userID,
		Username:        username,
		TransactionType: transactionType,
		TargetType:      targetType,
		Amount:          amount,
		BeforeAmt:       beforeAmt,
		AfterAmt:        afterAmt,
		Enable:          true,
		CreateBy:        "System",
		CreateDate:      time.Now(),
		ModifiedBy:      "",
		ModifiedDate:    time.Time{},
	}

	if db.TestMode {
		// Mock implementation for test mode
		db.mockTransactions = append(db.mockTransactions, &transaction)
		log.Printf("📊 [TEST] Created transaction for user %s (%s): Type: %d, Amount: %d, Before: %d, After: %d",
			userID.Hex(), username, transactionType, amount, beforeAmt, afterAmt)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.TransactionCollection.InsertOne(ctx, transaction)
	if err != nil {
		log.Printf("❌ Failed to create transaction for user %s: %v", userID.Hex(), err)
		return err
	}

	log.Printf("📊 Created transaction for user %s (%s): Type: %d, Amount: %d, Before: %d, After: %d",
		userID.Hex(), username, transactionType, amount, beforeAmt, afterAmt)
	return nil
}

func (db *Database) AccruePoints(userID primitive.ObjectID, username string, points int) error {
	wallet, err := db.GetUserWallet(userID)
	if err != nil {
		return err
	}

	beforeAmt := decimal128ToInt(wallet.Balance)

	_, err = db.UpdateWalletBalance(userID, points)
	if err != nil {
		return err
	}

	afterAmt := beforeAmt + points

	// Create transaction record (credit = 2, point accrual = 1)
	if db.TestMode {
		log.Printf("💰 [TEST] Awarded %d points to user %s (%s) - Balance: %d",
			points, username, userID.Hex(), afterAmt)
	} else {
		log.Printf("💰 Awarded %d points to user %s (%s)", points, username, userID.Hex())
	}
	return db.CreateTransaction(
		userID,
		username,
		2,
		1,
		points,
		beforeAmt,
		afterAmt,
	)

}
