# Design choices

## Database

The first choice I had to make was what kind of a backend to use. The options were:
 - PostgreSQL
 - SQLite
 - JSON
 - YAML
 
This is mean to be a small terminal app, so anything like Postgres is definitely an overkill. 

SQLite has a few advantages over pure JSON, including:
 - smaller file sizes (even down to 60% of JSON equivalent)
 - indexes improve query performance (given we don't query the WHOLE database)
 - we don't need to load the whole data into RAM

However, JSON also has its pros:
 - faster write times
 - faster read times (for reading WHOLE database)
 - easier to track undo/redo just by tracking changes in objects

Also, I can't imagine sending a `UPDATE`/`INSERT` and waiting for a response. The latency is simply too high. The only option is to make the same modification twice; in SQL query and in the local object, and run the updates in a transaction on save.

Sure, that's not a bad idea for a bigger project. But for a small CLI app like this one, it might be better to swallow the disadvantages of using JSON for faster dev times and less bugs. Besides, I ran a simulation and even if you complete ~30 tasks a day for 10 years, you are unlikely to exceed 50-60MB worth of JSON. Make that 100MB, and with a modern SSD it will take <2s for the initial load. I doubt anyone will be using this program for more than a few years in the first place, including me. Even if, and they want to keep the file sizes lower, they can just use a different file after 10 years if they care about slightly lower loading times and the file sizes.

YAML? Could use, but AFAIK JSON parsers are faster and for sure used much more frequently (therefore more field-tested).