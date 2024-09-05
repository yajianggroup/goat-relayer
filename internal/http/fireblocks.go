package http

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/goatnetwork/goat-relayer/internal/config"
	"github.com/golang-jwt/jwt/v5"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

const (
	webhookPkProduction = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA0+6wd9OJQpK60ZI7qnZG
jjQ0wNFUHfRv85Tdyek8+ahlg1Ph8uhwl4N6DZw5LwLXhNjzAbQ8LGPxt36RUZl5
YlxTru0jZNKx5lslR+H4i936A4pKBjgiMmSkVwXD9HcfKHTp70GQ812+J0Fvti/v
4nrrUpc011Wo4F6omt1QcYsi4GTI5OsEbeKQ24BtUd6Z1Nm/EP7PfPxeb4CP8KOH
clM8K7OwBUfWrip8Ptljjz9BNOZUF94iyjJ/BIzGJjyCntho64ehpUYP8UJykLVd
CGcu7sVYWnknf1ZGLuqqZQt4qt7cUUhFGielssZP9N9x7wzaAIFcT3yQ+ELDu1SZ
dE4lZsf2uMyfj58V8GDOLLE233+LRsRbJ083x+e2mW5BdAGtGgQBusFfnmv5Bxqd
HgS55hsna5725/44tvxll261TgQvjGrTxwe7e5Ia3d2Syc+e89mXQaI/+cZnylNP
SwCCvx8mOM847T0XkVRX3ZrwXtHIA25uKsPJzUtksDnAowB91j7RJkjXxJcz3Vh1
4k182UFOTPRW9jzdWNSyWQGl/vpe9oQ4c2Ly15+/toBo4YXJeDdDnZ5c/O+KKadc
IMPBpnPrH/0O97uMPuED+nI6ISGOTMLZo35xJ96gPBwyG5s2QxIkKPXIrhgcgUnk
tSM7QYNhlftT4/yVvYnk0YcCAwEAAQ==
-----END PUBLIC KEY-----`
	webhookPkSandbox = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAw+fZuC+0vDYTf8fYnCN6
71iHg98lPHBmafmqZqb+TUexn9sH6qNIBZ5SgYFxFK6dYXIuJ5uoORzihREvZVZP
8DphdeKOMUrMr6b+Cchb2qS8qz8WS7xtyLU9GnBn6M5mWfjkjQr1jbilH15Zvcpz
ECC8aPUAy2EbHpnr10if2IHkIAWLYD+0khpCjpWtsfuX+LxqzlqQVW9xc6z7tshK
eCSEa6Oh8+ia7Zlu0b+2xmy2Arb6xGl+s+Rnof4lsq9tZS6f03huc+XVTmd6H2We
WxFMfGyDCX2akEg2aAvx7231/6S0vBFGiX0C+3GbXlieHDplLGoODHUt5hxbPJnK
IwIDAQAB
-----END PUBLIC KEY-----`
)

// UNCHECKED
type FireblocksWebhookRequest struct {
	Type      string      `json:"type"`
	TenantId  string      `json:"tenantId"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// TRANSACTION_CREATED, TRANSACTION_STATUS_UPDATED, TRANSACTION_APPROVAL_STATUS_UPDATED
type TransactionDetails struct {
	ID                            string                   `json:"id"`
	ExternalTxID                  string                   `json:"externalTxId"`
	Status                        string                   `json:"status"`
	SubStatus                     string                   `json:"subStatus"`
	TxHash                        string                   `json:"txHash,omitempty"`
	Operation                     string                   `json:"operation"`
	Note                          string                   `json:"note"`
	AssetID                       string                   `json:"assetId"`
	AssetType                     string                   `json:"assetType"`
	Source                        TransferPeerPathResponse `json:"source"`
	SourceAddress                 string                   `json:"sourceAddress,omitempty"`
	Destination                   TransferPeerPathResponse `json:"destination"`
	Destinations                  []DestinationsResponse   `json:"destinations,omitempty"`
	DestinationAddress            string                   `json:"destinationAddress,omitempty"`
	DestinationAddressDescription string                   `json:"destinationAddressDescription,omitempty"`
	DestinationTag                string                   `json:"destinationTag,omitempty"`
	AmountInfo                    AmountInfo               `json:"amountInfo"`
	TreatAsGrossAmount            bool                     `json:"treatAsGrossAmount"`
	FeeInfo                       FeeInfo                  `json:"feeInfo"`
	FeeCurrency                   string                   `json:"feeCurrency"`
	CreatedAt                     int64                    `json:"createdAt"`
	LastUpdated                   int64                    `json:"lastUpdated"`
	CreatedBy                     string                   `json:"createdBy"`
	SignedBy                      []string                 `json:"signedBy,omitempty"`
	RejectedBy                    string                   `json:"rejectedBy,omitempty"`
	ExchangeTxID                  string                   `json:"exchangeTxId,omitempty"`
	CustomerRefID                 string                   `json:"customerRefId,omitempty"`
	ReplacedTxHash                string                   `json:"replacedTxHash,omitempty"`
	NumOfConfirmations            int                      `json:"numOfConfirmations"`
	BlockInfo                     BlockInfo                `json:"blockInfo"`
	Index                         int                      `json:"index,omitempty"`
	SystemMessages                []SystemMessageInfo      `json:"systemMessages,omitempty"`
	AddressType                   string                   `json:"addressType,omitempty"`
	RequestedAmount               float64                  `json:"requestedAmount,omitempty"`
	Amount                        float64                  `json:"amount,omitempty"`
	NetAmount                     float64                  `json:"netAmount,omitempty"`
	AmountUSD                     float64                  `json:"amountUSD,omitempty"`
	ServiceFee                    float64                  `json:"serviceFee,omitempty"`
	NetworkFee                    float64                  `json:"networkFee,omitempty"`

	// NetworkRecords               []NetworkRecord               `json:"networkRecords,omitempty"`
	// AuthorizationInfo            AuthorizationInfo             `json:"authorizationInfo"`
	// AmlScreeningResult           AmlScreeningResult            `json:"amlScreeningResult"`
	// ExtraParameters              TransactionExtraParameters    `json:"extraParameters"`
	// SignedMessages               []SignedMessage               `json:"signedMessages,omitempty"`
	// RewardsInfo                  RewardsInfo                   `json:"rewardsInfo"`
}

type AmountInfo struct {
	Amount          string `json:"amount"`
	RequestedAmount string `json:"requestedAmount"`
	NetAmount       string `json:"netAmount"`
	AmountUSD       string `json:"amountUSD"`
}

type FeeInfo struct {
	NetworkFee string `json:"networkFee"`
	ServiceFee string `json:"serviceFee"`
}

type TransferPeerPathResponse struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	SubType string `json:"subType"`
}

type DestinationsResponse struct {
	Amount                        string                   `json:"amount"`
	Destination                   TransferPeerPathResponse `json:"destination"`
	AmountUSD                     float64                  `json:"amountUSD"`
	DestinationAddress            string                   `json:"destinationAddress"`
	DestinationAddressDescription string                   `json:"destinationAddressDescription,omitempty"`
	CustomerRefID                 string                   `json:"customerRefId,omitempty"`

	// AmlScreeningResult         AmlScreeningResult      `json:"amlScreeningResult"`
}

type SystemMessageInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type BlockInfo struct {
	BlockHeight string `json:"blockHeight"`
	BlockHash   string `json:"blockHash"`
}

// handleFireblocksWebhook process webhook callback of Fireblocks, should validate request data first
func handleFireblocksWebhook(c *gin.Context) {
	if err := verifyWebhookSig(c); err != nil {
		log.Errorf("Fireblocks webhook sig verify error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sig error"})
		return
	}
	var req FireblocksWebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Infof("Fireblocks webhook received, Type: %s, TenantId: %s, Timestamp: %d, Data: %v",
		req.Type, req.TenantId, req.Timestamp, req.Data)

	// TODO: record
	if req.Type == "TRANSACTION_CREATED" {
		if trans, ok := req.Data.(TransactionDetails); ok {
			log.Infof("Fireblocks webhook transaction created detect, txId %s, txHash %s, status %s", trans.ExternalTxID, trans.TxHash, trans.Status)
		} else {
			log.Error("Fireblocks webhook failed to convert data to TransactionDetails")
		}
	}
	if req.Type == "TRANSACTION_STATUS_UPDATED" {
		if trans, ok := req.Data.(TransactionDetails); ok {
			log.Infof("Transaction status updated detect, txId %s, txHash %s, status %s, subStatus %s, assetType %s, amount %s", trans.ExternalTxID, trans.TxHash, trans.Status, trans.SubStatus, trans.AssetType, trans.AmountInfo.Amount)
			if trans.Operation == "TRANSFER" && trans.Status == "COMPLETED" && trans.SubStatus == "CONFIRMED" {
				// TODO: withdraw filter (sourceAddress, destinationAddress, assetType)
				log.Infof("Fireblocks webhook transaction confirmed, txId %s, txHash %s, sourceAddress %s, destinationAddress %s, assetType %s, amount %s", trans.ExternalTxID, trans.TxHash, trans.SourceAddress, trans.DestinationAddress, trans.AssetType, trans.AmountInfo.Amount)
			}
		} else {
			log.Error("Fireblocks webhook failed to convert data to TransactionDetails")
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleFireblocksCosignerTxSign process cosigner tx sign callback of Fireblocks, it will check tx and sign
func handleFireblocksCosignerTxSign(c *gin.Context) {
	if config.AppConfig.FireblocksPrivKey == "" || config.AppConfig.FireblocksPubKey == "" {
		log.Error("Cosigner callback empty RSA key")
		c.String(http.StatusInternalServerError, "Private key and public key not exist")
		return
	}

	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	rawBody := string(bodyBytes)

	// Parse and verify JWT
	tx, err := jwt.Parse(rawBody, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrInvalidKey
		}
		return config.AppConfig.FireblocksPubKey, nil
	})
	if err != nil || !tx.Valid {
		log.Error("Cosigner callback JWT valid false")
		c.Status(http.StatusUnauthorized)
		return
	}

	// Extract requestId from the JWT claims
	claims, ok := tx.Claims.(jwt.MapClaims)
	if !ok || !tx.Valid {
		log.Error("Cosigner callback JWT claims parsing failed")
		c.Status(http.StatusUnauthorized)
		return
	}

	requestId := claims["requestId"].(string)
	txId := claims["txId"].(string)

	log.Infof("Cosigner callback JWT claim received, requestId %s, txId %s", requestId, txId)

	// TODO: check by more fields

	// Sign the response APPROVE|REJECT|RETRY
	action := "APPROVE"
	rejectionReason := ""

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"action":          action,
		"requestId":       requestId,
		"rejectionReason": rejectionReason,
	})
	signedRes, err := token.SignedString(config.AppConfig.FireblocksPrivKey)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, signedRes)
}

// verifyWebhookSig verify sig from webhook request
func verifyWebhookSig(c *gin.Context) error {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	// Get sig
	signature := c.GetHeader("fireblocks-signature")
	if signature == "" {
		return errors.New("signature missing")
	}

	// Decode from base64
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	// Extract pk
	block, _ := pem.Decode([]byte(webhookPkProduction))
	if block == nil {
		return errors.New("failed to parse public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return errors.New("invalid public key")
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}

	// Create SHA-512 hash
	hash := sha512.New()
	hash.Write(body)
	hashed := hash.Sum(nil)

	// Verify sig
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA512, hashed, sig)
	if err != nil {
		log.Errorf("Fireblocks webhook verification failed: %v", err)
		return errors.New("invalid signature")
	}
	return nil
}
