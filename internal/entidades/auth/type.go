package auth

type CreateUserParams struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type LoginUserParams struct{
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
