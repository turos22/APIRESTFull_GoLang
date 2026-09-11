package auth

type CreateUserParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type LoginUserParams struct{
	Email    string `json:"email"`
	Password string `json:"password"`
}
