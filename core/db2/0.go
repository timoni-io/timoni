package db2

import (
	"core/config"
	"core/db2/sqlite2"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lukx33/lwhelper/out"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

var dbConnection = func() *gorm.DB {
	// Connecto to DB
	// conn, err := gorm.Open(postgres.Open("host="+POSTGRES_HOST+" user="+POSTGRES_USER+" password="+POSTGRES_PASS+" dbname="+POSTGRES_DB+" port="+POSTGRES_PORT+" sslmode=disable"), &gorm.Config{})
	conn, err := gorm.Open(sqlite2.Open(filepath.Join(config.DataPath(), "core.db")), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		// panic("failed to connect database: " + err.Error())
		fmt.Println("failed to connect database: " + err.Error())
		os.Exit(1)
	}

	conn.AutoMigrate(&certS{})
	conn.AutoMigrate(&certProviderS{})
	conn.AutoMigrate(&dNSProviderS{})
	conn.AutoMigrate(&domainS{})
	conn.AutoMigrate(&eventS{})
	conn.AutoMigrate(&imageBuilderS{})
	conn.AutoMigrate(&kubeS{})
	conn.AutoMigrate(&moduleStateS{})
	conn.AutoMigrate(&settingsS{})

	conn.AutoMigrate(&gitProviderS{})
	conn.AutoMigrate(&ingressS{})
	conn.AutoMigrate(&metricsS{})
	conn.AutoMigrate(&organizationS{})
	conn.AutoMigrate(&notificationProviderS{})

	return conn
}()

var dbLock = &sync.Mutex{}

type req_listQueryS struct {
	Where  string
	Order  string
	Offset int
	Limit  int
}

// ---

func generateTotpSecret(issuer, email string) string {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
		// Secret:      generateRandomB32Secret(12),
	})
	if err != nil {
		out.New(err)
		return ""
	}

	return key.Secret()
}

// off:
// func generateTOTP(email string) (qr []byte, secret string) {
// 	key, err := totp.Generate(totp.GenerateOpts{
// 		Issuer:      Config.Name,
// 		AccountName: email,
// 		// Secret:      generateRandomB32Secret(12),
// 	})

// 	if err != nil {
// 		log.Error(err)
// 		return nil, ""
// 	}

// 	img, _ := key.Image(200, 200)
// 	b := bytes.NewBuffer(nil)
// 	png.Encode(b, img)

// 	return b.Bytes(), key.Secret()
// }
