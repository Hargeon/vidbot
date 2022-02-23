package videocmprs

const (
	Initial    = "initial"
	Attributes = "attributes"
	Video      = "video"
)

type Account struct {
	ID int64 `jsonapi:"primary,tg_accounts"`

	ChatID int64 `jsonapi:"attr,chat_id"`
	// the token uses for registration
	Token string `jsonapi:"attr,token"`
	// the token uses for authentication
	TokenAuth string `jsonapi:"attr,token_auth,omitempty" json:"token_auth"`

	State       string   `json:"state"`
	Attributes  []string `json:"attr"`
	CurrentAttr string

	Request *Request `json:"request"`
}
