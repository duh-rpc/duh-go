> This README is a work in progress, and will likely be in a separate repo separating the spec from the implementation.

## Content Negotiation
This spec defines support for only the following mime types which can be specified in the `Content-Type` or `Accept` headers. However, since this is HTTP, service Implementors are free to add support for whatever mime type they want.

* `application/json` - This MUST be used when sending or receiving JSON. The charset MUST be UTF-8
* `application/protobuf` - This MUST be used when sending or receiving Protobuf. The charset MUST be ascii
* `application/octet-stream` - The MUST be used when sending or receiving unstructured binary data. The charset is undefined. The client/server should receive and store the binary data in it's unmodified form.
* `text/plain` - This should NOT be returned by service implementations. If the response has this content type or has no content type, this indicates to the client the response is not from the service, but from the HTTP infra or some other part of the HTTP stack that is outside of the service implementations control.

> The service implementation MUST always return the content type of the response. It MUST NOT return `text/plain`.

#### Content-Type and Accept Headers
Clients SHOULD NOT specify any mime type parameters as specified in RFC2046.  Any parameters after `;` in the provided content type will be ignored by the server.

The Content Type is expected to be specified in the following format, omitting any mime type parameters like `;charset=utf-8` or `;q=0.9`.
```
Content-Type: <MIME_type>/<MIME_subtype>
```


> Multiple mime types are not allowed, `Content-Type` and `Accept` header MUST NOT contain multiple mime types separated by comma as defined in [RFC7231](https://www.rfc-editor.org/rfc/rfc7231#section-5.3.2) If multiple mime types are provided, the server will ignore any mime type beyond the first provided.
>
> Implementations that add new mime types are encouraged to also follow this rule as it simplifies client and server implementations.

The mime types supported can change depending on the method. This allows service implementations to migrate from older mime types or to support mime types of a specific use case.

If none of the mime types can be accommodated by the server, the server WILL return code `400` and error in the format `Accept header 'application/bson' is invalid format or unrecognized content type, only [application/json, application/protobuf] are supported by this method`

> 
> Server MUST ALWAYS support JSON, if no other content type can be negotiated, then the server will always respond with JSON.

## Replies
Standard replies from the service SHOULD follow a common structure. This provides a consistent and simple method for clients to reply with errors and messages.
#### Reply Structure
The reply structure has the following fields.
* **Code** (Optional)  - Machine Readable Integer
* **Message** - A Human Readable Message
* **Details** - (Optional) An map of details about this error. Could include a link to the documentation explaining this error OR more machine readable codes and types.

> 
> Although the **Reply** structure is typically used for error replies, it can be used in normal `200` responses when there is a desire to avoid adding a new `<MethodCall>MessageResponse`  for simple method calls which have no detailed responses.
### Errors
Errors are returned using the Reply structure where the HTTP status code and the **Code** in the reply structure are always the same. This makes it clear the service is responding with the code and not some intermediate proxy or other infrastructure.
#### HTTP Status Codes
The HTTP status code MUST match the `Code` provided in the **Reply**  structure in the body of the reply if the HTTP Status code is NOT `200`

Service Implementations SHOULD only return the following standard HTTP Status codes along with the **Reply** structure defined above.

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
HTTP Status Codes should NOT be expanded for your specific use case. Instead server implementations should add their own custom fields and codes in the standard `v1.Reply.Details` map.

For example, a credit card processing company needs to return card processing errors. The recommended path would be to add those `details` to the standard error.
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
It is NOT recommended to add your own custom fields to the **Reply** structure. This approach would require clients to use a non standard structure. For example, the following is not recommened.
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
All Server responses MUST ALWAYS return JSON if no other content type is specified. It WILL NOT return `text/plain`. This includes any and all router based errors like `Method Not Allowed` or `Not Found` if the route requested is not found. This is because ambiguity exists between a route not found on the service, and a `Not Found` route at the API gateway or Load Balancer.

The server should always respond with a standard **Reply** structure which differentiates it's responses from any infrastructure that lies between the service and the client.

The Client SHOULD handle responses that do not include the **Reply** structure and differentiate those responses to as to clearly differentiate between a service `Not Found` replies and infrastructure `Not Found` responses. Implementations of the client SHOULD assume the response is from the infrastructure if it receives a reply that does NOT conform to the **Reply** structure.

For example, if the client implementation recieves an HTTP status code of `404` and a status message of `Not Found` from the request, the client SHOULD assume the error is from the infrastructure and inform the caller in a way that is suitable for the language used.
##### Service Identifiers
In addition, the server CAN include the `Server: DUH-RPC/1.0 (Golang)` header according to [RFC9110](https://www.rfc-editor.org/rfc/rfc9110#field.server) to help identify the source of the HTTP Status. (It is possible that proxy or API gateways will scrub or over write this header as a security measure, which will make identification of the source more difficult) 
