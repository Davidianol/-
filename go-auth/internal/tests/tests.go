package test

import (
	"Auth/internal/db"
	"Auth/internal/hash"
	"context"
	"fmt"
	"time"

	// iteraptor "Auth/internal/interaptors"
	// "Auth/internal/server"

	log "github.com/sirupsen/logrus"
)

func formatting(i *int, text string) {
	*i += 1
	fmt.Printf("TEST %s: %d\n", text, *i)
	fmt.Println("----------------------")
}

func RunTestsNewUser() {
	time.Sleep(time.Second)
	i := 0
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := db.InitDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	hash, err := hash.GetHashPassword("Password")
	if err != nil {
		log.Fatal(err)
	}
	user := db.User{
		Login:    "Terw",
		Password: hash,
		Role:     "CLIENT",
		Phone:    "+79999999999",
		Mail:     "gmail@maul.ru",
	}
	formatting(&i, "New user add")
	userr, err := db.AddUser(ctx, user)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(userr)
	formatting(&i, "Arleady user add")
	_, err = db.AddUser(ctx, user)
	if err != nil {
		log.Error(err)
	}
	formatting(&i, "Get data from user")
	out, err := db.GetProfileUserDataByLogin(ctx, user.Login)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(out)
}
