/*
Chatus is a go library for wechat public account(微信公共平台). It handles a few basic
wechat interface.

Functions
With it, you can setup a web server easily to do the following
- Handle different formats of requests (messages/events);
- Customizable handler to process different types requests;


Usage
- Extend handler to add your own business logic
- go build chatus
- ./chatus -s STATION_ID -w WEBSITE -t TOKEN


Files
- message.go
 -- parse JSON/XML messages sent from wechat;
 -- repackage the message into its corresponding struct;
 -- convert struct into JSON/XML messages that can be sent to wechat server
   -- Pull:  response to a user's request
   -- Push:  message to user

- chatus.go
 -- web api

*/
package chatus
