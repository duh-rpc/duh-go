> This README is a work in progress, and will likely be in a separate repo separating the spec from the
> implementation. 
> 
> PLEASE FEEL FREE TO OPEN A PR AND CORRECT OR ADD ANYTHING.

# DUH-RPC - Duh, Use Http for RPC
The goal of this project is not to duplicate GRPC or any other RPC style framework. It is in fact the opposite. It is an 
attempt to evangelize Standard HTTP for RPC. To help share the knowledge that you can get or exceed GRPC performance and 
consistency by simply using HTTP and following some basic rules and using a basic set of HTTP codes.

We started this journey after we realized [GRPC was slower than HTTP/1]. Next, we wanted to know if we could
gain the simplicity of an RPC framework, while exceeding performance and also reducing complexity. We realized, you
don't need a fancy framework to build performant and scalable APIs. All you need are some best practices born out of 
experience in building and scaling high-performance APIs that developers love.

## What is DUH-RPC?
The goal of this project isn't to create a new framework. The goal instead is to get developers thinking about WHY
they might be reaching for an RPC style framework, when what they actually want is a basic definition of naming, retry,
error and streaming semantics. This spec is an attempt to define those basic semantics which can be followed and
audited using standard HTTP and OpenAPI tooling. 

This repo includes a simple implementation of the DUH-RPC spec in golang to illustrate how easy it is to create a high
performance and scalable RPC style service by following the DUH-RPC spec.

DUH-RPC design is intended to be 100% compatible with OpenAPI tooling, Linters and Governance tooling to aid in the 
development of APIs Without compromising on Error Handling Consistency, Performance or Scalability.

# DUH Examples
### Successful Call
`POST http://localhost:8080/v1/say.hello {"name": "John Wick"}`
```json
{
    "message": "Hello, John Wick"
}
```
TODO MORE EXAMPLES HERE

### What's good about DUH-RPC?
* Avoid the hierarchy problem that come with using REST best practices.
* Any client that follows this spec can easily be adapted to your service or API.
* Consistent Error Handling, which allows libraries and frameworks to handle errors uniformly.
* Consistent RPC method naming, no need to second guess where in the hierarchy the operation should exist.
* Keeps the good parts of REST. Stateless, Cachable, Intermediates, Security
* The API can be interrogated from the command line via curl without the need for a special client.
* The API can be inspected and called from GUI clients like [Postman](https://www.postman.com/),
  or [hoppscotch](https://github.com/hoppscotch/hoppscotch)
* Use standard schema linting tools and OpenAPI-based services for integration and compliance testing of APIs
* Design, deploy and generate documentation for your API using standard Open API tools
* Use your favorite web frameworks in your favorite language.
* Consistent client interfaces allow for a set of standard tooling to be built to support common use cases.
  Like `retry` and authentication.
* Payloads can be encoded in any format (Like ProtoBuf, MessagePack, Thrift, etc...)
* Standard Streaming semantics (FUTURE WORK)
* You can use the same endpoints and frameworks for both private and public facing APIs, with no need to have separate
  tooling for each.

## When is DUH-RPC not appropriate?
DUH-RPC, like most RPC APIs, is mostly intended for service to service communication. It doesn't make sense to use
DUH-RPC if your intended use case is in a browser when users are sharing links to be clicked.

IE: https://www.google.com/search?q=rpc+api

## Why not GRPC?
Like DUH, GRPC has consistent semantics like flow control, request cancellation, and error handling. However, it is not without
its issues.
* GRPC is more complex than is necessary for high-performance, distributed environments.
* GRPC implementations can be slower than expected (Slower than standard HTTP)
* Using GRPC can result in more code than using standard HTTP
* GRPC is not suitable for the public Web based APIs

For a deeper dive and benchmarks of GRPC with standard HTTP in golang See [Why not GRPC](why-not-grpc.md)

## Why not REST?
Many who embrace RPC style frameworks do so because they are fleeing REST either because of the simple semantics of RPC
or for performance reasons. In our experience REST is suboptimal for a few reasons.
* The hierarchical nature of REST does not lend itself to clean interfaces over time.
* REST performance will always be slower than RPC
* No standard error semantics
* No standard streaming semantics
* No standard rate limit or retry semantics

For a deeper dive on REST See [Why not REST](why-not-rest.md)

# DUH Spec
> This spec is a work in progress and is not an exhaustive set of DO's and DON'Ts. If you have a suggestion, please
> consider opening a PR to help contribute to the spec.

Here we are presenting a standard approach for implementing RPC over HTTP. The benefit of using tried-and-true nature of HTTP
for RPC allows designers to leverage the many tools and frameworks readily available to design, document, and implement
high-performance HTTP APIs without overly complex client or deployment strategies.

TODO: discuss the separation of infra and service implementation. With DUH it is always clear to the client that the error
came from either infra, or the service. (unlike regular HTTP where you are not sure where the 404 came from)

DUH method calls take the form `/v1/my/endpoint.method`

### Requests
All requests SHOULD use the POST verb with an in body request object describing the arguments of the RPC request.
Because each request includes a payload in the encoding of your choice (Default to JSON) there is no need to use 
any other HTTP verb other than POST.

> In fact, if you attempt to send a payload using verbs like GET, You will find that some language frameworks assume 
> there is no payload on GET and will not allow you to read the payload.

The name of the RPC method should be in the standard HTTP path such that it can be easily versioned and routed by 
standard HTTP handling frameworks, using the form `/<version>/<subject>.<action>`

CRUD Examples
* `/v1/users.create`
* `/v1/users.delete`
* `/v1/users.update`
  Where `.` denotes that the action occurs on the `users` collection.

Methods SHOULD always reference the thing they are acting upon
* `/v1/dogs.pet`
* `/v1/dogs.feed`

If Methods do not act upon a collection, then you should indicate the problem domain the method is acting on
* `/v1/messages.send`
* `/v1/trash.clear`
* `/v1/kiss.give`

Naming it `/v1/kiss.give` instead of just `/v1/kiss` is an important future proofing action. In the case you want to
add other methods to `/v1/kiss` you have a consistent API. For instance if you only had `/v1/kiss` then when you 
wanted to add the ability to `blow` a kiss, your api would look like

* `/v1/kiss` - Create a Kiss
* `/v1/kiss.blow` - Blow a Kiss

Instead of the more consistent
* `/v1/kiss.give` - Give a Kiss
* `/v1/kiss.blow` - Blow a Kiss

### Subject before Noun or Action
You may have noticed that every endpoint has the subject before the action. This is intentional and is useful for 
future proofing your API such that you may want to add more actions in the future. Just remember, if in doubt, design
your API like Yoda would speak.

### Versioning
DUH-RPC calls employ standard HTTP paths to make RPC calls. This approach makes versioning those methods easy and direct.
This spec does NOT have an opinion on the versioning semantic used. IE: `v1` or `V1` or `v1.2.1`. It is highly recommended
that all methods are versioned in some way.

## Content Negotiation
This spec defines support for only the following mime types which can be specified in the `Content-Type` or `Accept` 
headers. However, since this is HTTP, service Implementors are free to add support for whatever mime type they want.

* `application/json` - This MUST be used when sending or receiving JSON. The charset MUST be UTF-8
* `application/protobuf` - This MUST be used when sending or receiving Protobuf. The charset MUST be ascii
* `application/octet-stream` - The MUST be used when sending or receiving unstructured binary data. The charset is 
  undefined. The client/server should receive and store the binary data in it's unmodified form.
* `text/plain` - This should NOT be returned by service implementations. If the response has this content type or has
  no content type, this indicates to the client the response is not from the service, but from the HTTP infra or some
  other part of the HTTP stack that is outside of the service implementations control.

> The service implementation MUST always return the content type of the response. It MUST NOT return `text/plain`.

#### Content-Type and Accept Headers
Clients SHOULD NOT specify any mime type parameters as specified in RFC2046.  Any parameters after `;` in the provided
content type will be ignored by the server.

The Content Type is expected to be specified in the following format, omitting any mime type parameters like
`;charset=utf-8` or `;q=0.9`.

```
Content-Type: <MIME_type>/<MIME_subtype>
```

> Multiple mime types are not allowed, `Content-Type` and `Accept` header MUST NOT contain multiple mime types separated
> by comma as defined in [RFC7231](https://www.rfc-editor.org/rfc/rfc7231#section-5.3.2) If multiple mime types are 
> provided, the server will ignore any mime type beyond the first provided or may optionally ignore the Content-Type
> completely and return JSON. 
>
> Implementations that add new mime types are encouraged to also follow this rule as it simplifies client 
> and server implementations as the RFC style of negotiation is unnecessary within the scope of RPC.

The mime types supported can change depending on the method. This allows service implementations to migrate from older 
mime types or to support mime types of a specific use case.

If the server can accommodate none of the mime types, the server WILL return code `400` and a standard reply structure
with the message  
```
Accept header 'application/bson' is invalid format or unrecognized content type, only 
[application/json, application/protobuf] are supported by this method
```

> Server MUST ALWAYS support JSON, if no other content type can be negotiated, then the server will always respond with JSON.

## Replies
Standard replies from the service SHOULD follow a common structure. This provides a consistent and simple method for 
clients to reply with errors and messages.

#### Reply Structure
The reply structure has the following fields.
* **Code** (Optional)  - Machine Readable Integer
* **Message** - A Human Readable Message
* **Details** - (Optional) A map of details about this error which could include a link to the documentation explaining
  this error OR more machine-readable codes and types.

> Although the **Reply** structure is typically used for error replies, it CAN be used in normal `200` responses when
> there is a desire to avoid adding a new `<MethodCall>Response` type for simple method call which has no 
> detailed responses.

### Errors
Errors are returned using the Reply structure where the HTTP status code and the **Code** in the reply structure 
are always the same. This makes it clear the service is responding with the code and not some intermediate proxy or 
other infrastructure.

#### HTTP Status Codes
The HTTP status code MUST match the `Code` provided in the **Reply**  structure in the body of the reply if the HTTP
Status code is NOT `200`.

Service Implementations SHOULD only return the following standard HTTP Status codes along with the **Reply** structure 
defined above.

| Code          | Short                | Retry | Long                                                                                  |
| ------------- | -------------------- | ----- | ------------------------------------------------------------------------------------- |
| 200           | OK                   | N/A   | Everything is fine                                                                    |
| 400           | Bad Request          | False | Server reports missing a required parameter, or malformed request object              |
| 401           | Unauthorized         | False | Not Authenticated, Who are you?  (AuthN)                                              |
| 403           | Forbidden            | False | You can't access this thing, it either doesn't exist, or you don't have authorization |
| 404           | Not Found            | False | The thing you where looking for was not found                                         |
| 409           | Conflict             | False | The request conflicts with another request                                            |
| 429           | Too Many Requests    | True  | Stop abusing our service. (See Standard RateLimit Responses)                          |
| 452           | Client Error         | False | The client returned an error without sending the request to the server                |
| 453           | Request Failed       | False | Request is valid, but failed. If no other code makes sense, use this one              |
| 500           | Internal Error       | True  | Something with our server happened that is out of our control                         |
| 501           | Not Implemented      | False | The method requested is not implemented on this server                                |

>  Most Standard HTTP Clients will handle 1xx and 3xx class errors, so it's not something you should need to worry about.

#### Errors and Codes
HTTP Status Codes should NOT be expanded for your specific use case. Instead server implementations should add their
own custom fields and codes in the standard `v1.Reply.Details` map.

For example, a credit card processing company needs to return card processing errors. The recommended path would be to
add those `details` to the standard error.
```json
{
	"code": 402,
	"message": "Credit Card was declined",
	"details": {
		"type": "card_error",
		"decline_code": "expired_card",
		"doc": "https://credit.com/docs/card_errors#expired_card"
	}
}
```
It is NOT recommended to add your own custom fields to the **Reply** structure. This approach would require clients to
use a non-standard structure. For example, the following is not recommended.
```json
{
	"code": 402,
	"message": "Credit Card was declined",
	"error": {
		"description": "expired_card",
		"department": {
			"id": "0x87DD0006",
			"sub_code": 1003002,
			"type": "E102"
		}
	},
	"details": {
		"doc": "https://credit.com/docs/80070045D/E102#1003002"
	}
}
```
#### Handling HTTP Errors
All Server responses MUST ALWAYS return JSON if no other content type is specified. It WILL NOT return `text/plain`. 
This includes any and all router based errors like `Method Not Allowed` or `Not Found` if the route requested is not
found. This is because ambiguity exists between a route not found on the service, and a `Not Found` route at the API
gateway or Load Balancer.

The server should always respond with a standard **Reply** structure which differentiates it's responses from any 
infrastructure that lies between the service and the client.

The Client SHOULD handle responses that do not include the **Reply** structure and differentiate those responses to
as to clearly differentiate between a service `Not Found` replies and infrastructure `Not Found` responses. 
Implementations of the client SHOULD assume the response is from the infrastructure if it receives a reply that does 
NOT conform to the **Reply** structure.

For example, if the client implementation receives an HTTP status code of `404` and a status message of `Not Found`
from the request, the client SHOULD assume the error is from the infrastructure and inform the caller in a way that 
is suitable for the language used.

### Infrastructure Errors
An infrastructure error is any HTTP response code that is NOT 200 and DOES NOT include a `Reply` structure in the body.
If the client receives a response code and it DOES NOT include a `Reply` structure in the expected serialization format,
then the client MUST consider the response as an infrastructure error and handling it accordingly.

Typically, infrastructure errors are 5XX class errors, but could also be 404 Not Found errors, or consist of 
non-standard or future HTTP status codes. As such it is recommended that client implementations do not attempt to 
handle all possible HTTP codes, but instead consider any non 200 responses without a `Reply` an infrastructure class
error.

##### Service Identifiers
In addition, the server CAN include the `Server: DUH-RPC/1.0 (Golang)` header according to [RFC9110](https://www.rfc-editor.org/rfc/rfc9110#field.server) to help
identify the source of the HTTP Status. (It is possible that proxy or API gateways will scrub or overwrite this 
header as a security measure, which will make identification of the source more difficult) 

### RPC Call Semantics
An RPC call or request over the network has the following semantics

### CRUD Semantics
Similar to REST semantics
`/subject.get` always returns data and does not cause side effects.
`/subject.create` creates a thing and can return the data it just created
`/subject.update` updates an existing resource
`/subject.delete` deletes an existing resource

##### Request reached service and the client received a response
The response from the service could be good or bad. The point is that were no interruptions occurred in order to 
impede the request from reaching the service which handles the request.

##### Request was rejected
The request could be rejected by the service or infrastructure for various reasons.
* IP Blacklisted
* Rate Limited
* No such Endpoint or Method found
* Not Authenticated
* Malformed or Incorrect request parameters
* Not Authorized

#### Request timed out
The request could have timed out by the infrastructure, client or service for various reasons.
* Upstream service timeout (DB or other service timeout)
* TCP idle timeout rule (proxy or firewall)
* 502 Gateway Timeout

##### Request is cancelled by the service
The request could have been cancelled by
* Service shutdown during processing
* Catastrophic Failure of the server or service

##### The caller cancels the request
The request timed out waiting for a reply, or the caller requested a cancel of the request while in flight.

##### Request was denied by infrastructure
The infrastructure attempted to connect the request with a service, but was denied
* 503 Service Unavailable (Load Balancer has no upstream backends to fulfill the request)
* Internal Error on the Load Balancer or Proxy
* TCP firewall drops SYN (Possibly due to SYN Flood, and silently never connects)

### RPC Call Characteristics
Based on the above possible outcomes we can infer that an RPC client should have the following characteristics.

##### Unary RPC Requests should be retried until canceled
Because a request could timeout, be rejected, be canceled, and denied for transient reasons, the client should
have some resiliency built-in and retry depending on the error received. The request should continue to retry with 
back off until the client determines the request took too long, or the client cancels the request.

TODO: FINISH
In order to support these characteristics, the service MUST reply with a well-defined set of error replies which 
the client can use to decide which operations should be retried and which should consistute a failure. Also,
the service MUST differentiate it's replies with replies from the infrastructure so the client will know what is 
appropriate to retry and what is not.

TODO: Not Found infra vs Not Found service, might mean we retry.

TODO: Streams

TODO: RateLimit responses (So clients can implement and retry mechanisms)

### FIN
If you got this far, go look at the `demo/client.go` and `demo/service.go` for examples of how 
to use this framework.

# DEMO

### Only POST is allowed
`GET http://localhost:8080/v1/say.hello`
```json
{
  "code": 400,
  "codeText": "Bad Request",
  "message": "http method 'GET' not allowed; only POST"
}
```

### Missing a request body
`POST http://localhost:8080/v1/say.hello`
```json
{
  "code": 454,
  "codeText": "Client Content Error",
  "message": "proto: syntax error (line 1:1): unexpected token "
}
```

### Validation error
`POST http://localhost:8080/v1/say.hello {"name": ""}`
```json
 {
  "code": 400,
  "codeText": "Bad Request",
  "message": "'name' is required and cannot be empty"
}
```



## Existing RPC Options
There are already plenty of frameworks to choose from.
* GPRC
* https://dubbo.apache.org
* [GRPC Web](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)
* [DRPC](https://github.com/storj/drpc)

