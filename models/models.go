package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserWallet represents the TblUserWallet collection structure
type UserWallet struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"UserID" json:"user_id"`
	WalletType   int                `bson:"WalletType" json:"wallet_type"`
	WalletName   string             `bson:"WalletName" json:"wallet_name"`
	Balance      int                `bson:"Balance" json:"balance"`
	Enable       bool               `bson:"Enable" json:"enable"`
	CreateBy     string             `bson:"CreateBy" json:"create_by"`
	CreateDate   time.Time          `bson:"CreateDate" json:"create_date"`
	ModifiedBy   string             `bson:"ModifiedBy" json:"modified_by"`
	ModifiedDate time.Time          `bson:"ModifiedDate" json:"modified_date"`
}

// TransactionMovement represents the TblTransactionMovement collection structure
type TransactionMovement struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"UserID" json:"user_id"`
	Username       string             `bson:"Username" json:"username"`
	TransactionType int               `bson:"TransactionType" json:"transaction_type"`
	TargetType     int                `bson:"TargetType" json:"target_type"`
	Amount         int                `bson:"Amount" json:"amount"`
	BeforeAmt      int                `bson:"BeforeAmt" json:"before_amt"`
	AfterAmt       int                `bson:"AfterAmt" json:"after_amt"`
	Enable         bool               `bson:"Enable" json:"enable"`
	CreateBy       string             `bson:"CreateBy" json:"create_by"`
	CreateDate     time.Time          `bson:"CreateDate" json:"create_date"`
	ModifiedBy     string             `bson:"ModifiedBy" json:"modified_by"`
	ModifiedDate   time.Time          `bson:"ModifiedDate" json:"modified_date"`
}


// AuthToken represents the authentication token structure
type AuthToken struct {
	UserID   primitive.ObjectID `json:"user_id"`
	Username string             `json:"username"`
	ExpiresAt time.Time         `json:"expires_at"`
}

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Remark         string             `bson:"Remark" json:"Remark"`
	Username       string             `bson:"Username" json:"Username"`
	Email          string             `bson:"Email" json:"Email"`
	Password       string             `bson:"Password" json:"Password"`
	Enable         bool               `bson:"Enable" json:"Enable"`
	CreateBy       string             `bson:"CreateBy" json:"CreateBy"`
	CreateDate     time.Time          `bson:"CreateDate" json:"CreateDate"`
	ModifiedBy     string             `bson:"ModifiedBy" json:"ModifiedBy"`
	ModifiedDate   time.Time          `bson:"ModifiedDate" json:"ModifiedDte"`
	UserType int `bson:"UserType" json:"UserType"`
	UserVip int `bson:"UserVip" json:"UserVip"`
	LastLoginIp string `bson:"LastLoginIp" json:"LastLoginIp"`
	ReferCode string `bson:"ReferCode" json:"ReferCode"`
	SessionToken string `bson:"SessionToken" json:"SessionToken"`
}
