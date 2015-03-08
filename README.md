# GoRedisVoter

Fast, concurrent voting application written with Go, using Redis as a database.
Easy API and optional web interface.

## Usage

There are 3 main things you can do with this application:
- Make a Ballot
- Query the Ballot
- Vote on a Ballot

## API
All parameters must be sent in the request body as JSON, not as part of a query string. 

### Make a Ballot: `Post` [appUrl]/make 
This is used to 
  
Example:
- URL: _localhost:3000/make_
- POST Body: _{"Name":"Dinner Time","Desc":"What should I have for dinner?","V1":"Pizza","V2":"Pasta","V3":"Burger","V4":"Fish","Time":600}_

### Query the Ballot: `Get` [appUrl]/ballot/[ballotID]
asld;okjkfh;ldasfj;ldasf

###Vote on a Ballot: `Post` [appUrl]/ballot/[ballotID]
asld;okjkfh;ldasfj;ldasf
