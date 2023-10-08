# Requests

# Live
 * EnvID: string
 * Limit: int
```json
{
    "RequestID": "123",
    "Action": "Live",
    "Args": {
        "Subs": [
            {
                "EnvID": "env-lav2zt0t1257",
                "Limit": 50,
                "Where": {
                    "Type": "IS",
                    "Value": {
                        "Field": "aa",
                        "Value": "vv"
                    }
                }
            },
            {
                "EnvID": "image-builder",
                "Limit": 50,
                "Where": {
                    "Type": "IS",
                    "Value": {
                        "Field": "aa",
                        "Value": "vv"
                    }
                }
            },
            {
                "EnvID": "traefik",
                "Limit": 50,
                "Where": {
                    "Type": "IS",
                    "Value": {
                        "Field": "aa",
                        "Value": "vv"
                    }
                }

            }
        ]
    }
}
```

# Get
```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Querys": [
            {
                "Query": {
                    "Type": "VECTOR",
                    "Direction": "AFTER",
                    "Time": 123,
                    "EnvID": "image-builder",
                    "Limit": 50,
                    "FullLog": true,
                    "Where": {
                        "Type": "OR",
                        "Value": [
                            {
                                "Type": "IS",
                                "Value": {
                                    "Field": "imageid",
                                    "Value": "alpine:pods-log-spamer.08e13b3fa6b634f55c9d147faa2847168eed8230"
                                }
                            }
                        ]
                    }
                }
            }
        ]
    }
}
```

# Query
## BASE
##### -- All query types inherits this type --
##### -- Not to be used on its own --
### EnvID: string
### Events: (true/FALSE)
### FullLog: (true/FALSE)
### Limit: int
### Where: Operator

---
## VECTOR
### Direction:
 * BEFORE
 * AFTER

### Time: uint64 (unix nano timestamp)

```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Query": {
            "Type": "VECTOR",
            "EnvID": "traefik",
            "Events": false,
            "Direction": "AFTER",
            "Time": 123,
            "Where": {}
        }
    }
}
```
---
## ONE
### Time: uint64 (unix nano timestamp)

```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Query": {
            "Type": "ONE",
            "Time": 123,
            "EnvID": "env",
            "FullLog": true,
            "Where": {
                "Type": "IS",
                "Value": {
                    "Field": "aa",
                    "Value": "vv"
                }
            }
        }
    }
}
```
---
## RANGE
### TimeBegin and TimeEnd: uint64 (unix nano timestamp)

```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Query": {
            "Type": "RANGE",
            "EnvID": "env",
            "TimeBegin": 123,
            "TimeEnd": 1234,
            "Where": {
                "Type": "IS",
                "Value": {
                    "Field": "aa",
                    "Value": "vv"
                }
            }
        }
    }
}
```
---
## TAGS
### Field:
 * empty - if listing all TAGS
 * string - if listing all TAG values

```json
{
    "RequestID": "123",
    "Action": "Get",
    "Args": {
        "Query": {
            "Type": "TAGS",
            "EnvID": "env",
            "Field": "",
            "Where": {
                "Type": "IS",
                "Value": {
                    "Field": "aa",
                    "Value": "vv"
                }
            }
        }
    }
}
```

# Operators
## AND
### Value: []Operator

```json
{
    "Type": "AND",
    "Value": [
        ...
    ]
}
```

---
## OR
### Value: []Operator

```json
{
    "Type": "OR",
    "Value": [
        ...
    ]
}
```

---
## NOT
### Value: Operator

```json
{
    "Type": "NOT",
    "Value": {
        ...
    }
}
```

---
## IS
### Value:
 - Field: string
 - Value: string


```json
{
    "Type": "IS",
    "Value": {
        "Field": "aaa",
        "Value": "123"
    }
}
```
---
## EXISTS
### Value:
 - Field: string

```json
{
    "Type": "EXISTS",
    "Value": {
        "Field": "aaa",
    }
}
```
---
## BETWEEN
### Value:
 - Field: string
 - From: (int/float)
 - To: (int/float)

```json
{
    "Type": "BETWEEN",
    "Value": {
        "Field": "aaa",
        "From": 123,
        "To": 1024
    }
}
```
