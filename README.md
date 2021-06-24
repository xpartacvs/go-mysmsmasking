# MySMSMasking

Package golang untuk berkomunikasi dengan API MySMSMasking

## Cara Pakai

Download paket:

```bash
go get github.com/xpartacvs/go-mysmsmasking
```

Langsung koding:

```go
package main

import (
    "github.com/xpartacvs/go-mysmsmasking"
    "fmt"
)

func main() {

    // Instansiasi client
    client := mysmsmasking.New("<url-server-api>", "<username>", "<password>")

    // Cek saldo dan kedaluarsa
    acc, err := client.GetAccountInfo()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Saldo\t\t: %d\n", acc.Balance)
    fmt.Printf("Kedaluarsa\t: %s\n\n", acc.Expiry.Format("2006-01-02 15:04:05 MST"))

    // Kirim SMS
    awb, err := client.Send("081xxxxxxxxxx", "Testing kiriman SMS dengan package golang MySMSMasking")
    if err != nil {
        panic(err)
    }
    fmt.Printf("ID Kiriman\t: %s\n", awb.Id)
    fmt.Printf("Waktu Kirim\t: %s\n\n", awb.Timestamp.Format("2006-01-02 15:04:05 MST"))

    // Cek status kiriman SMS
    status, err := client.GetReport(awb.Id)
    if err != nil {
        panic(err)
    }
    switch status {
    case mysmsmasking.DELIVERED:
        fmt.Printf("Status ID %s: SAMPAI", awb.Id)
    case mysmsmasking.SENT:
        fmt.Printf("Status ID %s: OTW", awb.Id)
    case mysmsmasking.INVALID_ID:
        fmt.Printf("Status ID %s: ID TIDAK KETEMU", awb.Id)
    default:
        fmt.Printf("Status ID %s: GAGAL", awb.Id)
    }

}
```
