package videocmprs

const (
	Initial      = "initial"
	Attributes   = "attributes"
	Video        = "video"
	Registration = "registration"
)

// Account uses for Telegram account
type Account struct {
	ID int64 `jsonapi:"primary,tg_accounts"`

	ChatID int64 `jsonapi:"attr,chat_id"`
	// Email uses only for registration telegram account
	Email string `jsonapi:"attr,email"`
	// TokenAuth uses for authentication
	TokenAuth string `jsonapi:"attr,token_auth,omitempty" json:"token_auth"`

	State       string   `json:"state"`
	Attributes  []string `json:"attr"`
	CurrentAttr string

	Request *Request `json:"request"`
}
