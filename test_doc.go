package main

import (
	"context"
	"fmt"
	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
	"net/http"
	"net/http/httptest"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"_index":"test","_type":"_doc","_id":"delete-1","found":false}`))
	}))
	defer ts.Close()

	c, _ := client.New(config.WithAddresses(ts.URL))
	
	getResp, err := builder.NewDocumentBuilder(c, "test").
		ID("delete-1").
		Get(context.Background())
	fmt.Printf("Get returned: getResp=%v, err=%v\n", getResp, err)
}
