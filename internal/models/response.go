package models

type Result struct {
	URL       string
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Response struct {
	Results []Result
}
