# JCrypt

Easily encrypt and decrypt annotated fields on-the-fly during JSON marshalling.

## Marshaling

Export any data type in JSON representation as done by `json.Marshal` and encrypt certain values:

```go
import (
    "fmt"

    "github.com/sbreitf1/go-jcrypt"
)

type data struct {
    UserName string `json:"username"`
    Password string `json:"password" jcrypt:"aes"`
}

func main() {
    raw, _ := jcrypt.Marshal(d, &jcrypt.Options{
        GetKeyHandler: jcrypt.StaticKey([]byte("secret")),
    })

    fmt.Println(string(raw))
}
```

The above example will output something like

```json
{
    "username":"obi wan",
    "password": {
        "mode":"aes",
        "data":"-uHW77tqZg8ATOVIApk9Wgh3C78x8NZl4E6xFWOTM-i1YsgKwi5NuGYOYNjg6t0pmBQawjxuRT7qDPyMaoGP1A"
    }
}
```

The `jcrypt` annotation causes the `password` field to be encrypted using AES. The corresponding encryption key is `secret` and is passed as static key.

## Unmarshaling

Obtaining the plaintext value from an encrypted JSON representation is also comparable to `json.Unmarshal`:

```go
import "github.com/sbreitf1/go-jcrypt"

type data struct {
    UserName string `json:"username"`
    Password string `json:"password" jcrypt:"aes"`
}

func main() {
    var d data
    jcrypt.Unmarshal(jsonInputData, &d, &jcrypt.Options{
        GetKeyHandler: jcrypt.StaticKey([]byte("secret")),
    })
}
```

## Missing Features

- Unmarshal other datatypes than string
- Encrypt arbitrary values (only strings can be encrypted at the moment)
- Intensive tests
- Respect json-annotation options like `omitempty` and `string`
- Document GetKey-Callback handler for interactive password input
- Other encryption standards
- Check for encryption / disable and document fallback-mode for unencrypted values in annotated fields
- Auto-encrypt files in fallback-mode
- YAML support