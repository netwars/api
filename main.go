package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/julienschmidt/httprouter"
	"github.com/netwars/api/cache"
	"github.com/piotrkowalczuk/rest"
	resthttprouter "github.com/piotrkowalczuk/rest/httprouter"
	"golang.org/x/net/context"
)

var (
	warmUp    int
	debugAddr string
	httpAddr  string
)

const (
	storageEntryExpiration     = 24 * time.Hour
	storageEntryInterval       = 30 * time.Second
	internalServerErrorMessage = "Oops... something goes wrong!"
	netwarsURL                 = "http://netwars.pl"
)

func main() {
	// Flag domain. Note that gRPC transitively registers flags via its import
	// of glog. So, we define a new flag set, to keep those domains distinct.
	fs := flag.NewFlagSet("", flag.ExitOnError)

	fs.IntVar(&warmUp, "warmup", 0, "number of pages per forum to fetch on start")
	fs.StringVar(&debugAddr, "debug.addr", ":8000", "Address for HTTP debug/instrumentation server")
	fs.StringVar(&httpAddr, "http.addr", ":8001", "Address for HTTP (JSON) server")

	flag.Usage = fs.Usage // only show our flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	u, err := url.Parse(netwarsURL)
	if err != nil {
		logger.Fatal(err)
	}

	client := NewClient(u)
	topicCache := cache.NewCache(cache.CacheOpts{
		Expiration: storageEntryExpiration,
		Interval:   storageEntryInterval,
	})
	topicStorage := NewTopicStore(client, topicCache, TopicStoreOpts{
		WarmUp: warmUp,
	})
	go logErrorChannel("topic-storage", topicStorage.Err())

	// Transport: HTTP (debug/instrumentation)
	go func() {
		logger.Fatal(http.ListenAndServe(debugAddr, nil))
	}()

	ctx := context.Background()
	ctx = NewTopicStorageContext(ctx, topicStorage)

	logger.Fatal(http.ListenAndServe(httpAddr, buildRoutes(ctx)))
}

func buildRoutes(ctx context.Context) *httprouter.Router {
	router := httprouter.New()
	router.GET("/topic/:topicId", buildHandler(ctx, TopicGetEndpoint, TopicGetRequestDecode))
	router.GET("/topics", buildHandler(ctx, TopicsGetEndpoint, TopicsGetRequestDecode))
	router.GET("/forums", buildHandler(ctx, ForumsGetEndpoint, nil))

	return router
}

func buildHandler(ctx context.Context, end endpoint.Endpoint, decode rest.Decode) httprouter.Handle {
	if decode == nil {
		decode = func(context.Context, *http.Request) (interface{}, error) {
			return nil, nil
		}
	}
	return resthttprouter.InjectParamsToContext(&rest.Server{
		Context: ctx,
		Endpoint: rest.GenerateRequestID(
			rest.BasicEndpointCancellation(
				end,
			),
		),
		DecodeFunc: decode,
		EncodeFunc: rest.JSONEncode,
		After:      []rest.After{},
		Before:     []rest.Before{},
		ErrorFunc: func(ctx context.Context, rw http.ResponseWriter, err error) {
			log.Println(err)

			switch e := err.(type) {
			case *rest.Error:
				http.Error(rw, e.Message, e.HTTPCode)
			default:
				http.Error(rw, internalServerErrorMessage, http.StatusInternalServerError)
			}

		},
	})
}

func logErrorChannel(prefix string, err <-chan error) {
	for e := range err {
		log.Printf("[%s] - %s", prefix, e.Error())
	}
}
