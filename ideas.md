Google Sheet (DONE)
-------------------

Send the spreadsheet to the server

```json
[["Logger"], []]
```

Server (DONE)
-------------

http://server => { Number: status }

We number all the houses from 1 to whatever

so block 1 house 1 = 1, block 1 house 2 = 2, etc

Website
-------

[ ] Get the positions of all the houses and make an object which goes
    Number => (x, y) on image

[ ] fetch() the data, loop through each one and draw a circle on the canvas if
            the point associated with that number is sold.
