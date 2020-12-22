package internal

import "log"

func requireNoErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
