> This README is a work in progress, and will likely be in a separate repo separating the spec from the
> implementation. It should be refactored to focus on the high-level benefits of DUH instead of spending the first
> half of the document GRPC bashing (I might have been angry) and then link to a separate document with the full spec.
> 
> PLEASE FEEL FREE TO OPEN A PR AND CORRECT OR ADD ANYTHING.

# DUH-RPC - Duh, Use Http for RPC

Here we are presenting a standard approach for implementing RPC over HTTP. The benefit of using tried-and-true nature of HTTP
for RPC allows designers to leverage the many tools and frameworks readily available to design, document, and implement
high performance HTTP APIs without overly complex client or deployment strategies.

The goal of this spec is not to duplicate GRPC or to make GPRC better. It is in fact the opposite. It is an attempt to
evangelize HTTP for RPC. To share the knowledge that you can get or exceed GRPC performance and consistency by 
simply using HTTP.

Many frame works have been proposed in an attempt to overcome or to compound the core problems with using GPRC. But 
this is not one of those. You don't need a fancy framework to make performant and scalable APIs.

The current list of GRPC or modified GRPC frameworks
* https://dubbo.apache.org
* [GRPC Web](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)
* [DRPC](https://github.com/storj/drpc)

## Why not GPRC?
While GRPC has a lot going for it, there are a few aspects that nudged us toward a different direction. Firstly, it
demands a language-specific framework, with service definitions in Protobuf. This can deter broad adoption, especially
for software as a service-based companies who wish to have customers use their APIs. Also, large interdepartmental
or business units have to create and provide APIs for integration and operation of their respective units. A strict 
GRPC interface ecosystem is only viable within an organization which is completely committed to its adoption, and 
has accepted that there will be one-off standard HTTP endpoints for external customers.

### GRPC is not suitable for the Web
As mentioned above, when adopting GRPC you accept that any public facing API's designed to be accessible via the WEB
at scale cannot be GRPC. GRPC is useful only as an internal communication method to make calls between services.

Much as been written about why GRPC is not suitable for the web, but the core rational is that GRPC uses sticky
connections, and it prefers client side load balancing. Both of which make load balancing clients difficult by
placing control in the hands of the client, not the API service provider.

* [GRPC Weaknesses via Microsoft](https://learn.microsoft.com/en-us/aspnet/core/grpc/comparison?view=aspnetcore-8.0#grpc-weaknesses)
* [When to avoid GRPC via Redhat](https://www.redhat.com/architect/when-to-avoid-grpc)
* [GRPC-Web requires a proxy](https://blog.envoyproxy.io/envoy-and-grpc-web-a-fresh-new-alternative-to-rest-6504ce7eb880)
* [GRPC Load Balancing via Microsoft](https://learn.microsoft.com/en-us/aspnet/core/grpc/performance?view=aspnetcore-8.0#load-balancing)

### GPRC is complex and often Slow
As you dive deeper into GRPC, you'll notice GRPC carries some baggage in the form of 
[deprecated systems and a somewhat bloated structure](https://www.storj.io/blog/introducing-drpc-our-replacement-for-grpc),
which can leave users scratching their heads on what's best to use and how. When performance suffers, it's not always 
clear why GPRC is doing what it's doing. For instance, we ran into a strange performance issue when using streaming 
which led us to discover that GRPC queue's requests once the concurrent number of streams has been hit. (Our only
recourse was to re-design the system with the knowledge of GPRCs nonsensical limitations.)
(See [GRPC Docs](https://grpc.io/docs/guides/performance/))

With the proliferation of HTTP2, the gap of performance between GRPC and HTTP1 has been greatly and in many cases
eclipsed by standard HTTP/2 implementation. Our own hands-on experiments showed that GRPC's heralded performance is 
not as great as it used to be, especially in high-concurrency, low-latency scenarios. In fact, the entire reason this
spec exists is the realization, that our standard HTTP services were outperforming our GRPC based services.

See https://github.com/duh-rpc/duh-go-benchmarks for a comparison of DUH with GRPC in golang. (prepare yourself for a shock)

### GRPC Gateway
TODO: Talk about the downsides and performance issues we have experienced using GRPC gateway.
* Multiple ports over loading the existing port
* TLS connection to local gateway issues. (See Gubernator)

## Why not REST?
The main reason REST will always be slower than GRPC and DUH-RPC is due to REST requiring an HTTP Router for
matching complex REST paths like `/v1/{thing}/collection/{id}/foo`. Both GRPC and DUH-RPC use simple string matching to
connect an RPC method to a handler which results in unparalleled performance.

RPC style API's also do not suffer from the hierarchical structure problems that rest APIs do. 
(TODO: Write a blog post about this and link it here) or if you are at mailgun, read the 'HTTP and GRPC Design Guide'
in the Mailgun wiki.

## Why HTTP?
* The widespread adoption of HTTP/2 has closed the gap in performance and features between other popular RPC style frameworks.
* No need to adopt a framework that may or may not exist in your language.
* You have access to an entire ecosystem of tools and services, IE: you can use CURL or GUI's to make API calls.
* Payloads can be encoded in any format (Like ProtoBuf, BSON, Thrift, etc...)
* You can use the same endpoints and frameworks for both private and public facing APIs, with no need to have separate 
  tooling for each.

### What's good about DUH-RPC?
* Avoid the hierarchy problems that come with using REST best practices (See TODO above)
* Any client that follows this spec can easily be adapted to your service.
* Consistent Error Handling, no need to guess if you are doing it correctly
* Consistent RPC method naming, no need to second guess where on the hierarchy you should place your operation.
* The API can be interrogated from the command line with curl without the need for a special client.
* The API can be inspected and called from GUI clients like [Postman](https://www.postman.com/), 
  or [hoppscotch](https://github.com/hoppscotch/hoppscotch)
* Use standard schema linting tools and OpenAPI-based services for integration and compliance testing of API's
* Design, deploy and generate documentation for your API using standard Open API tools
* Use your favorite web frameworks in your favorite language.
* Consistent client interfaces allow for a set of standard tooling to be built to support common use cases. 
  Like `retry` and authentication.

DUH-RPC design is intended to be 100% compatible with OpenAPI tooling, Linters and Governance tooling to aid in the 
development of APIs Without compromising on Error Handling Consistency, Performance or Scalability.

# DUH Implementation
TODO: Write some gentle introduction here instead of dropping right into the spec. Give some examples and responses
provide some examples of handling errors and such so people can visually see what this is all about. Also discuss
the separation of infra and service implementation. With DUH it is always clear to the client that the error
came from either infra, or the service. (unlike regular HTTP where you are not sure where the 404 came from)

DUH method calls take the form `/v1/my/endpoint.method`

### Requests
All requests SHOULD use the POST verb with an in body request object describing the arguments of the RPC request.
Because each request includes a payload in the encoding of your choice (Default to JSON) there is no need to use 
any other HTTP verb other than POST.

> In fact, if you attempt to send a payload using verbs like GET, You will find that some language frameworks assume 
> there is no payload on GET and will not allow you to read the payload.

The name of the RPC method should be in the standard HTTP path such that it can be easily versioned and routed by 
standard HTTP handling frameworks, using the form `/<version>/<domain>.<action>`

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
| 402           | Request Failed       | False | Request is valid, but failed. If no other code makes sense, use this one              |
| 403           | Forbidden            | False | You can't access this thing, it either doesn't exist, or you don't have authorization |
| 409           | Conflict             | False | The request conflicts with another request                                            |
| 428           | Client Error         | False | The client returned an error without sending the request to the server                |
| 429           | Too Many Requests    | True  | Stop abusing our service. (See Standard RateLimit Responses)                          |
| 500           | Internal Error       | True  | Something with our server happened that is out of our control                         |
| 501           | Not Implemented      | False | The method requested is not implemented on this server                                |
| 502, 503, 504 | Infrastructure Error | True  | Something is wrong with the infrastructure                                            |

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

##### Service Identifiers
In addition, the server CAN include the `Server: DUH-RPC/1.0 (Golang)` header according to [RFC9110](https://www.rfc-editor.org/rfc/rfc9110#field.server) to help
identify the source of the HTTP Status. (It is possible that proxy or API gateways will scrub or over write this 
header as a security measure, which will make identification of the source more difficult) 

### RPC Call Semantics
An RPC call or request over the network has the following semantics

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

##### Request is cancelled by the caller
The request timed out waiting for a reply or the caller requested a cancel of the request while in flight.

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
In order to support these characteristics, the service MUST reply with a well defined set of error replies which 
the client can use to decide which operations should be retried and which should consistute a failure. Also,
the service MUST differentiate it's replies with replies from the infrastructure so the client will know what is 
appropriate to retry and what is not.

TODO: Not Found infra vs Not Found service, might mean we retry.

TODO: Streams

TODO: RateLimit responses (So clients can implement standard fall back and retry mechanisms.)

### FIN
If you got this far, go look at the `demo/client.go` and `demo/service.go` for examples of how 
to use this framework.

