package models


type Config struct {
	DbURI   string `json:"connectUriDb"`
	LogName string `json:"logName"`
	Port    string `json:"port"`
}	

type Account struct{
	Id int `gorm:"id"`
	Login string `gorm:"column:login"`
	Pass string `gorm:"column:pass"`
}

func (Account) TableName() string {
	return "account"
}

type Users struct {
	Id int `gorm:"id"`
	Login string `gorm:"login"`
	Password string `gorm:"password"`
	Status string `gorm:"status"`
	ExitDate string `gorm:"exitdate"`
}

func (Users) TableName() string {
	return "users"
}

type Message struct {
	UserId int `json:"user_id"`
	User string `json:"user"`
	Message string `json:"message"`
	Date string `json:"date"`
}

type Logs struct {
	Id int `gorm:"column:id"`
	User string `gorm:"column:user"`
}

func (Logs) TableName() string {
	return "logs"
}

type History struct {
	Id int `gorm:"column:id"`
	User string `gorm:"column:user"`
	Message string `gorm:"column:message"`
	Date string `gorm:"column:date"`
}

func (History) TableName() string {
	return "history"
}
