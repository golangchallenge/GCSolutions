# mosaic2go - Go Challenge 3

http://golang-challenge.com/go-challenge3/

This app requires a memcached server for session handling and S3 for file storage,
set the appropriate config in the config.json file. I planed to run it on Heroku,
therefore the need for S3 file storage and Memcached. The AWS S3 client requires a
credention file as described here http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html.
The tmp folder needs write permission.

Build and run the application with:

$ go build -o mosaic
$ ./mosaic

Run the tests:

$ go test ./...

Access it in the browser: http://localhost:8080

Or online at Heroku: https://mosaic2go.herokuapp.com/


This is a REST based mosaic generator API. Unfortunately there as not enough time
to build the frontend, but you call it with curl and generate a mosaic.

Following endpoints are supported:

POST /user/anonymous
Create a new anonymous user session that lasts for 30 minutes.

GET /user/info
Returns the current user session.

GET /mosaic
Returns the last created mosaic.

POST /mosaic
Creates a mosaic for the current target image with tile images for query parameter "q".

POST /mosaic/target
Uploads a target image

GET /mosaic/target
Returns the current target image

GET /mosaic/tiles
Searches and returns tile images for query parameter "q".


Step by stepguide to generate a mosaic:

Step 1:

Initialize a user session. The returned authToken must be set as header for all further requests
$ curl -XPOST localhost:8080/user/anonymous

Step2:

Upload a target image:
$ curl -XPOST -H "authToken:[REPLACE WITH AUTH TOKEN]" -F "image=@/absolute/path/to/image.jpg" localhost:8080/mosaic/target

Step3:

Generate the mosaic with cat images:
$ curl -XPOST -H "authToken:[REPLACE WITH AUTH TOKEN]" localhost:8080/mosaic?q=cats
