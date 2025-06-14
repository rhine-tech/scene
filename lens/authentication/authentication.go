package authentication

type User struct {
	// UserID unique id. should be generated by server.
	UserID string `json:"user_id" bson:"user_id"`
	// Username unique name. should be unique. should be input by user. but can be changed
	Username string `json:"username" bson:"username"`
	// Password just password. for now, store in plain text
	Password string `json:"password" bson:"password"`
	// Email user's email address
	Email string `json:"email" bson:"email"`
}

type UserInfo struct {
	// Stored value
	UserID      string `json:"user_id" bson:"user_id"`
	DisplayName string `json:"display_name" bson:"display_name"`
	Avatar      string `json:"avatar" bson:"avatar"`
	Timezone    string `json:"timezone" bson:"timezone"`
	// Derived value
	Username string `json:"username" bson:"username,omitempty"` // from User
	Email    string `json:"email" bson:"email,omitempty"`       // from User
}

type LoginStatus struct {
	UserID   string `json:"user_id"`
	Token    string `json:"token"`
	Verifier string `json:"verifier"`
	ExpireAt int64  `json:"expire_at"`
}
