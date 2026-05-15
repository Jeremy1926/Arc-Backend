# auth

## Plans

For all public requests, they should be sent in the api group /api/v1/auth

### /api/v1/auth/session POST

What we should plan on doing is having communication with the issuer, either through some rpc or socket of some sort.

Identity format:

```go
{
"iat": int, // Issued at
"exp": int, // Expiration (should be 240)
"iss": string, // Issuer
"sub": string, // Account Identifier
"dn": string, // Display Name
"cty": string, // Country
}
```

Okay, since we finished the base creation on our service aswell as the hashing, we now need to get onto the part where we actually use this. I'm not sure where i want to start with communication yet.
