package auth

type User struct {
	UserID   int
	Name     string
	Email    string
	Password string

	Key string
}

type KeyPair struct {
	Key    string
	Secret string
}
