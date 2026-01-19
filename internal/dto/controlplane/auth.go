package controlplane

// ---------- Login ----------

type LoginRequest struct {
	Username string `json:"username,omitempty" binding:"required"`
	Phone    string `json:"phone,omitempty" binding:"required"`
	Email    string `json:"email,omitempty" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn: access token validity duration in seconds
	ExpiresIn int64 `json:"expires_in"`
}

// ---------- Register ---------
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type RegisterResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// ----------- Logout --------------------

type LogoutResponse struct {
	Success bool `json:"success"`
}

// ---------- Me ----------

type MeResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ---------- Refresh ----------

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	// ExpiresIn: access token validity duration in seconds
	ExpiresIn int64 `json:"expires_in"`
}

// ---------- Change Password ----------
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
