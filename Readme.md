## Chat server/client exercise using go

### Server
* Run: go run server/main/server.go
* usage: &#60;command&#62; [arguments]
    * help
    * exit
    * name              &#60;my name&#62;
    * group-add       &#60;group name&#62;      &#60;user name&#62;
    * group-rm        &#60;group name&#62;      &#60;user name&#62;
    * ls-user
    * ls-group        [group name]
    * sendmsg         &#60;uer name&#62;         &#60;message&#62;
    * sendgmsg       &#60;group name&#62;      &#60;message&#62;
    * sendbmsg       &#60;message&#62;

### Client
* Run: go run client/main/client.go &#60;host:port&#62;