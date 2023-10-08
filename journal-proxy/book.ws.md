```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Query": {
            "Type": "MULTI",
            "Queries": [
                {       
                    "Type": "ONE",
                    "EnvID": "traefik",
                    "Time": 1667221085000029906
                },
                {       
                    "Type": "ONE",
                    "EnvID": "traefik",
                    "Time": 1667210861299284434
                }
            ]
        }
    }
}
```

```json
{
    "RequestID": "123",
    "Action": "Live",
    "Args": {
        "Subs":[{
            "EnvID": "dev",
            "Limit": 20
        }
        ],
        "Limit":20

    }
}
```

```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Querys": [
            {
                "Query": {
                    "Type": "VECTOR",
                    "Time": "1691480492834091596",
                    "EnvID": "dev",
                    "Direction": "AFTER",
                    "Limit": 20
                }
            }
        ]
    }
}
```
