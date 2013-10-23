package heroku_test

import (
	"log"

	"github.com/kr/heroku-go"
)

func Example() {
	c := &heroku.Client{Host: heroku.DefaultHost}

	// create an app
	app := &heroku.App{}
	err := c.Create(app)
	if err != nil {
		log.Fatal(err)
	}

	// get app info
	app = &heroku.App{Name: "myapp"}
	err = c.Info(app)
	if err != nil {
		log.Fatal(err)
	}

	// update an app
	err = c.Update(&heroku.App{Name: "myapp", Maintenance: true})
	if err != nil {
		log.Fatal(err)
	}

	// delete an app
	err = c.Destroy(&heroku.App{Name: "myapp"})
	if err != nil {
		log.Fatal(err)
	}

	var apps []*heroku.App
	err = c.List(&apps)
	if err != nil {
		log.Fatal(err)
	}
}
