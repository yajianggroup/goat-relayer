package http

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestFireblocksCosignerTxSignJwt(t *testing.T) {
	// Create a valid JWT token
	bodyBytes := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJ0eElkIjoiZDEwYzE0NTgtZjM1Mi00NTg4LWJmMDgtMTMzYzIzYzZkZDA4Iiwib3BlcmF0aW9uIjoiVFJBTlNGRVIiLCJzb3VyY2VUeXBlIjoiVkFVTFQiLCJzb3VyY2VJZCI6IjEiLCJkZXN0VHlwZSI6Ik9ORV9USU1FX0FERFJFU1MiLCJkZXN0SWQiOiIiLCJhc3NldCI6IkVUSF9URVNUNSIsImFtb3VudCI6MC4wMDEwMDAwMCwiYW1vdW50U3RyIjoiMC4wMDEiLCJyZXF1ZXN0ZWRBbW91bnQiOjAuMDAxMDAwMDAsInJlcXVlc3RlZEFtb3VudFN0ciI6IjAuMDAxIiwiZGVzdEFkZHJlc3MiOiIweDI4QTMxMzQzZDdkNzAwYzRDQTUwYkJFMjIzNTU1ZjU3YzlhOTc1MDYiLCJleHRyYVBhcmFtZXRlcnMiOnsic2VjdXJpdHlFbnJpY2htZW50Ijp7fX0sImRlc3RpbmF0aW9ucyI6W3siYW1vdW50TmF0aXZlIjowLjAwMTAwMDAwLCJhbW91bnROYXRpdmVTdHIiOiIwLjAwMSIsImFtb3VudFVTRCI6Mi4zNzIwMzQ4NiwiZHN0QWRkcmVzcyI6IjB4MjhBMzEzNDNkN2Q3MDBjNENBNTBiQkUyMjM1NTVmNTdjOWE5NzUwNiIsImRzdEFkZHJlc3NUeXBlIjoiT05FX1RJTUUiLCJkc3RJZCI6IiIsImRzdFRhZyI6ImxldCdzIGdvb28iLCJkc3RUeXBlIjoiT05FX1RJTUVfQUREUkVTUyIsImRpc3BsYXlEc3RBZGRyZXNzIjoiMHgyOEEzMTM0M2Q3ZDcwMGM0Q0E1MGJCRTIyMzU1NWY1N2M5YTk3NTA2IiwiZGlzcGxheURzdFRhZyI6ImxldCdzIGdvb28iLCJhY3Rpb24iOiIyLVRJRVIiLCJhY3Rpb25JbmZvIjp7ImNhcHR1cmVkUnVsZU51bSI6MCwicnVsZXNTbmFwc2hvdElkIjo0NzE1NiwiYnlHbG9iYWxQb2xpY3kiOmZhbHNlLCJieVJ1bGUiOnRydWUsInJ1bGVUeXBlIjoiVEVOQU5UIiwiY2FwdHVyZWRSdWxlIjoie1wiZHN0XCI6e1wiaWRzXCI6W1tcIipcIl1dfSxcInNyY1wiOntcImlkc1wiOltbXCIxXCIsXCJWQVVMVFwiLFwiKlwiXV19LFwidHlwZVwiOlwiVFJBTlNGRVJcIixcImFzc2V0XCI6XCIqXCIsXCJhY3Rpb25cIjpcIjItVElFUlwiLFwiYW1vdW50XCI6MCxcIm9wZXJhdG9yc1wiOntcIndpbGRjYXJkXCI6XCIqXCJ9LFwicGVyaW9kU2VjXCI6MCxcImFtb3VudFNjb3BlXCI6XCJTSU5HTEVfVFhcIixcImFtb3VudEN1cnJlbmN5XCI6XCJVU0RcIixcImRzdEFkZHJlc3NUeXBlXCI6XCJPTkVfVElNRVwiLFwiYXBwbHlGb3JBcHByb3ZlXCI6ZmFsc2UsXCJ0cmFuc2FjdGlvblR5cGVcIjpcIlRSQU5TRkVSXCIsXCJhbGxvd2VkQXNzZXRUeXBlc1wiOlwiRlVOR0lCTEVcIixcImV4dGVybmFsRGVzY3JpcHRvclwiOlwie1xcXCJpZFxcXCI6XFxcImQ0MDMwZmQ5LWZiY2UtNGE3Ny05NjNjLWQwZTIwNTljZmZkMVxcXCJ9XCIsXCJhdXRob3JpemF0aW9uR3JvdXBzXCI6e1wibG9naWNcIjpcIk9SXCIsXCJncm91cHNcIjpbe1widGhcIjoyLFwidXNlcnNcIjpbXCJhMjQzMjlhYy1mYjVhLTQ3ZGQtYmJjMC1mMTI1M2I4MzUxNDhcIl0sXCJ1c2Vyc0dyb3Vwc1wiOltcIjE3YzhmYTY5LTQ4MDUtNDlhMS04YTE1LTFlOTdmYzA1YTUwMlwiXX1dLFwiYWxsb3dPcGVyYXRvckFzQXV0aG9yaXplclwiOnRydWV9fSJ9LCJhdXRob3JpemF0aW9uR3JvdXBzIjp7ImxvZ2ljIjoiT1IiLCJncm91cHMiOlt7InRoIjoyLCJ1c2VycyI6WyJhMjQzMjlhYy1mYjVhLTQ3ZGQtYmJjMC1mMTI1M2I4MzUxNDgiXSwidXNlcnNHcm91cHMiOlsiMTdjOGZhNjktNDgwNS00OWExLThhMTUtMWU5N2ZjMDVhNTAyIl19XSwiYWxsb3dPcGVyYXRvckFzQXV0aG9yaXplciI6dHJ1ZX19XSwibm90ZSI6IkdPQVQgTDIgU0QhIiwicmVxdWVzdElkIjoiZDEwYzE0NTgtZjM1Mi00NTg4LWJmMDgtMTMzYzIzYzZkZDA4Iiwic2lnbmVySWQiOiJiMzJkY2VmNC02OWU4LTQyZTctYWFmYi0yZjU4ZjkyZDViZmUifQ.aNQw-Kzj4wh3-7aO01li09QxEyuT0-EitPJOJFx2sJQSYaPq29FdoRN1eFQKhZlJTVDVyovJvghuUH87thyM8MUEgZL6v9g9XUefCVANyE8ZaY1u1ADyM0QUzcuevQonE_-986zry7RAzQqzgiGM_4vv9-9HEAdHNrYV4nL88olBjghqtP5aaaL9Qajtalm2m_7YLOVFvuGh02xN4NYAPpmLREwZGDMb5anBXY93rCEmSGsfQPMf44-qmSjBMTlHh0-vuPFXBiHD69I6EScx-R4c6VFUk3Aqx8vML2rIMl8smwQ4v4byTfXfJ9hMwFG0kUYLJGZ2oZkVznKS44sFmQ"
	cosignerPubKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsp4b/tDJfke0LtWoQmBn
HEPevpL7hR5QwbwmqRSuQN8eU/nCWZRi5K7gy09fwgPQxaGy8868+VqkJBXRuENS
5GizTWArX/NeO0ERdPzxGU9YUqWjC5wmz4il9AjEkhZ32kEEEt8FiY5SwCvaa2a0
8h7VVD2AKi+IlVwHhzOm9VXwTNbckO+JFilWET+/6SoQnSnWTld458jYKMD3tTTs
obwpt4QjfL+lNlfjkbrqhLHWA1pUKLfFaDG4LdmtOe9uC7lcwnO4kiyLI2ppC0mZ
mzDC2REwx8Kq4c5UnhwnSaUUQ/3JEnbGe0EPCImmb9TyFXPszJzHolC5Yf0nz2yh
BQIDAQAB
-----END PUBLIC KEY-----`
	rawBody := string(bodyBytes)

	block, _ := pem.Decode([]byte(cosignerPubKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		t.Errorf("failed to parse PEM block containing the public key")
		return
	}

	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Errorf("failed to parse public key: %v", err)
		return
	}

	pubKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		t.Error("not an RSA public key")
		return
	}

	// Parse and verify JWT
	tx, err := jwt.Parse(rawBody, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrInvalidKey
		}
		return pubKey, nil
	})
	assert.NoError(t, err)
	assert.True(t, tx.Valid)

	// Extract requestId from the JWT claims
	claims, ok := tx.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	requestId := claims["requestId"].(string)
	txId := claims["txId"].(string)

	t.Logf("Cosigner callback JWT claim received, requestId %s, txId %s", requestId, txId)
}

func TestFireblocksWebhookData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// TODO args cannot be nil
	server := NewHTTPServer(nil, nil, nil)
	router.POST("/api/fireblocks/webhook", server.handleFireblocksWebhook)

	rawJson := `{"type":"TRANSACTION_CREATED","tenantId":"c6d2ba75-71bc-4587-8ceb-dbf1388b7139","timestamp":1725612415366,"data":{"id":"d2180931-ee56-4d63-851f-9bffcfa35f0c","createdAt":1725612274925,"lastUpdated":1725612274959,"assetId":"BTC_TEST","source":{"id":"","type":"UNKNOWN","name":"External","subType":""},"destination":{"id":"1","type":"VAULT_ACCOUNT","name":"goattest1","subType":""},"amount":0.00001,"networkFee":0.00001,"netAmount":0.00001,"sourceAddress":"tb1qqkwwqeraapk0jekl53jk5zznp6u0yemjalqk6e","destinationAddress":"tb1qysxt7h98c60z77wz0hpwkr793pfmptln88afef","destinationAddressDescription":"","destinationTag":"","status":"CONFIRMING","txHash":"b8fb5ecf4fe2f9de5d2aff07cd860cac853d32373ff2ece75f20255905583fc3","subStatus":"PENDING_BLOCKCHAIN_CONFIRMATIONS","signedBy":[],"createdBy":"","rejectedBy":"","amountUSD":0.56,"addressType":"","note":"","exchangeTxId":"","requestedAmount":0.00001,"feeCurrency":"BTC_TEST","operation":"TRANSFER","customerRefId":null,"numOfConfirmations":0,"amountInfo":{"amount":"0.00001","requestedAmount":"0.00001","netAmount":"0.00001","amountUSD":"0.56"},"feeInfo":{"networkFee":"0.00001"},"destinations":[],"externalTxId":null,"blockInfo":{"blockHash":null},"signedMessages":[],"index":0,"assetType":"BASE_ASSET"}}`
	reqBody := bytes.NewBufferString(rawJson)
	req, err := http.NewRequest(http.MethodPost, "/api/fireblocks/webhook", reqBody)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("fireblocks-signature", "f1hpfT3JDXzLDfcpqN4Ql8iYPYvweh+zma7zZt9COSgCL96l+dxxhNop+MiSHkDJbxyJSIfAOdzPzLLB4uQpP+7njpVF7SRlmuI7iW7mshMEgQeTAGmyzMDom9shUs9uDkVGRg6Khp7u8f6wsp5HcIfKI9IU4+s95Akkkk0inzWiNrYdkm3rduUKcaG2rJQ/L6MT+2sOKOtBmc6lDsamK7zAcRFTLh4kazVuUiPPHl07zF4qgbGOCMStzApBKsI7R82zD6dVvVNb/ne88KJySR70v+LWTySWgUkzjhJscyY3vk6gO46nyD2gazAhuwxyaxEPnSFNVW0WXUTT7vPEUEbX5Ct/ltD7E7/yed0XR1Z3WFfDmVOHdGP5WsGsBI/Bzm4R5TZXtfMq4MwvLH6p68CkpakTE0wRupLaoOPMPszUQtcj2XEFQOZ1eQM3+r8MvbuCzBMmxiZ3c8p8EpTE4ZQObCAMMkIwGdb6tLLrAQ6KxVx0Qbnxvhr6/S0tru5JLuE9A2fc5VAcmQpVDe3DrNoVyM8KbXVFuKU2KZwEa4ZCTM+fPTKBaEfn9FFXVSksieNF0HIIoNbRcBceo/DfU6R6KJ/6zP09+Pnw4J4Y83RK2sAoVokf+bbxHUFYeYu91foq0h7nQLbTRz+sidPwavJpkVsaKeZ2vvJRbZxdXgc=")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	t.Logf("response: %s", w.Body.String())

	expectedResponse := `{"status":"ok"}`
	assert.Equal(t, expectedResponse, w.Body.String())
}
