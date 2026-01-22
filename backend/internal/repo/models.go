package repo

type UserProfile struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	DOB       string `json:"dob"`
	Avatar    *string `json:"avatar_path"`
	Nickname  *string `json:"nickname"`
	About     *string `json:"about"`
	IsPublic  bool   `json:"is_public"`
	CreatedAt string `json:"created_at"`
}

type Post struct {
	ID        int64
	UserID    int64
	GroupID   *int64
	Text      string
	Visibility string
	MediaPath *string
	CreatedAt string
}

type Comment struct {
	ID        int64
	PostID    int64
	UserID    int64
	Text      string
	MediaPath *string
	CreatedAt string
}

type Group struct {
	ID          int64
	CreatorID   int64
	Title       string
	Description string
	CreatedAt   string
}

type Notification struct {
	ID        int64
	UserID    int64
	Type      string
	Payload   string
	IsRead    bool
	CreatedAt string
}

type Message struct {
	ID        int64
	FromID    int64
	ToID      *int64
	GroupID   *int64
	Text      string
	CreatedAt string
}
