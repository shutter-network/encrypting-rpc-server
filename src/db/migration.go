package db

import (
	"fmt"

	"github.com/shutter-network/encrypting-rpc-server/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

const BufferSize = 10

type PostgresDb struct {
	DB          *gorm.DB
	addTxCh     chan TransactionDetails
	inclusionCh chan TransactionDetails
}

type TransactionDetails struct {
	Address         string `gorm:"primaryKey;index:idx_address_nonce"`
	Nonce           uint64 `gorm:"primaryKey;index:idx_address_nonce"`
	TxHash          string `gorm:"primaryKey;index:idx_tx_hash"`
	EncryptedTxHash string `gorm:"primaryKey"`
	SubmissionTime  int64
	InclusionTime   uint64
	IsCancellation  bool
}

func InitialMigration(dbUrl string) (*PostgresDb, error) {

	gormConfig := &gorm.Config{Logger: gorm_logger.Default.LogMode(gorm_logger.Silent)}

	db, err := gorm.Open(postgres.Open(dbUrl), gormConfig)
	if err != nil {
		utils.Logger.Error().Err(err).Msg("failed to connect database")
		return nil, fmt.Errorf("failed to connect database | err: %v", err)
	}

	// run migrations
	if err := db.AutoMigrate(TransactionDetails{}); err != nil {
		utils.Logger.Error().Err(err).Msg("failed to automigrate tables")
		return nil, fmt.Errorf("failed to automigrate tables | err: %v", err)
	}

	inclusionCh := make(chan TransactionDetails, BufferSize)
	addTxCh := make(chan TransactionDetails, BufferSize)

	return &PostgresDb{DB: db, addTxCh: addTxCh, inclusionCh: inclusionCh}, nil
}
