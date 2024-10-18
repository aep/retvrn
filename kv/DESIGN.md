given 1 million documents with 100 attributes,
some of which are long text

```graphql

type Org {
    id:         ID!
    name:       String!
    address:    [String!]
    Phones:     [Phone!]
    Contacts:   [Contact!]
}

type Phone {
    id:     ID!
    Number: String!
}

type Contact {
    id:      ID!
    Name:    String!
    Phones:  [Phone!]
}

```

ideally we'd use a column family for each index, but that's not supported by tikv.

the values are stored in index 0 with field name, then id, then array index or map key

    f . name     . $id1       -> ACME
    f . address  . $id1 . 1   -> Main Street
    f . address  . $id1 . 2   -> Big City
    f . name     . $id2       -> Bob
    f . number   . $id3       -> 123
    f . name     . $id4       -> Motör


a sort and search index is generated from name and the normalized value.
the lookup is fuzzy, so any filter needs to re-rerun against the forward index if we want a strict match.
long text can be split into multiple entries and stemmed

    r . name  . acme   . T  . $id1
    r . name  . motor  . T  . $id2
    r . name  . bob    . T  . $id3

unordered relationships (egdes)

    g . contacts . $id1 . $id2
    g . phones   . $id1 . $id3


with both indexes we can build most queries now:

 - WHERE name = ACME                | s . name . acme
 - WHERE contacts != NULL 		    | i . contacts . .
 - WHERE number  > 2                | s . number . 2
 - WHERE comment contains "robot" 	| s . comment  . robot
 - ORDER BY number                  | s . number .
 - SELECT name WHERE type = "org" 	| type  . . org  . \$ ->  name . \$


