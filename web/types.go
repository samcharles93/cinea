package web

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

type MediaItem struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Overview  string `json:"overview"`
	PosterURL string `json:"poster_url"`
}
