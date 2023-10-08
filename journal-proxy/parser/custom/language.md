# Parser Language
    {{tag1}} - {{tag2}}. {{tag3}}

Log Line:

    Hello - World. Message

Parsed:
```json
{"tag1": "Hello", "tag2": "World", "tag3": "Message"}
```

## Tag
{{ tag_name | parameter | list }}

 * tag_name - only letters and '_'
 * parameter_list - supported paramaters split with '|' char

## Supported parameters

 * include
 * trim
 * trim_left
 * trim_right

### include number
includes multiple parts into tag

    {{tag | include 2 }}. {{msg}}

Log Line: 

    Hello.World.!!!. Message

Parsed:
```json
{"tag": "Hello.World.!!!", "msg": "Message"}
```
### trim 'charset'
trims left and right chars

    {{tag | trim '!' }}. {{msg}}

Log Line: 

    !Hello!!. Message

Parsed:
```json
{"tag": "Hello", "msg": "Message"}
```