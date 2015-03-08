# GoRedisVoter

Fast, concurrent voting application written with Go, using Redis as a database.
Easy API and in development web interface.

## Usage

There are 3 main things you can do with this application:

- Make a Ballot
- Query the Ballot
- Vote on a Ballot

Ballots are given up to 4 voting options, a short name, a long description and an end time. Before the end time is met people can place votes on any of the 4 voting options or "None of the above" option. After the end time is met, voting closes and the result is made available. Each IP address can only vote once and the finished result is available for 24 hours after the Ballot closes, at which time it is deleted.

## API
All parameters must be sent in the request body as JSON, not as part of a query string. 

### Make a Ballot: `Post` [appUrl]/make 
This is used to make a Ballot. Feed it a JSON document in the HTTP request body with the following parameters:

| Name | Value Description | Value Example |
| ---- | ---- | ---- |
| "Name" | Display the help window | "Dinner Time" |
| "Description" | Closes a window | "What should I have for dinner tonight?" |
| "V1" | Vote catagory 1 | "Pizza" |
| "V2" | Vote catagory 2 | "Pasta" |
| "V3" | Vote catagory 3| "Burger" |
| "V4" | Vote catagory 4 | "Fish" |
| "Time" | Time(s) for ballot to be open | 600 |
  
Example:
- URL: _localhost:3000/make_
- POST Body: _{"Name":"Dinner Time","Desc":"What should I have for dinner?","V1":"Pizza","V2":"Pasta","V3":"Burger","V4":"Fish","Time":600}_
- Response: 303 redirect to localhost:3000/ballot/*BallotID*

### Query the Ballot: `Get` [appUrl]/ballot/[ballotID]
This API request is used to query the ballot. if the ballot is still active and accepting votes, you will only recieve back the ballot name, description and vote catagory names. Once the vote is over you will also get back the vote results. This requires no JSON in the body of the request, just the ballot ID in the URL.


###Vote on a Ballot: `Post` [appUrl]/ballot/[ballotID]
asld;okjkfh;ldasfj;ldasf
