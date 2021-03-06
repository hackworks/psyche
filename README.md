## Psyche the mind

**It is a weekend project, beware before completely relying on it**

Psyche is a smart backend service for bots. It exposes endpoints for performing message relay, index messages based on hash tags with smart tag extraction to enrich and simple search.


### Plugins

All endpoints receive message via HTTP POST. Different features are implemented as plugins. Every endpoint has a dedicated plugin to handle the request. Chaining of plugins is not implemented yet.


#### Relay `/relay`

Messaging based collaboration typically requires the user to join multiple chat rooms or channels or topics to keep up with what is happening around. It is difficult to keep hopping between rooms and respond to messages that require your attention, especially if you are not tagged.

The `relay` plugin allows you to listen to messages in different chat rooms and send them to a room of your choice. You will now have a single pane view of things happening around you.


#### Register `/register`

For the `relay` plugin to work, we need to provide a room and POST endpoint. We use `botler` for getting messages and posting responses.

The `register` plugin allows end users to provide the mapping of rooms and POST endpoints. These mappings are stored persistently in the `Psyche` service.


#### Indexer `/indexer`

Lack of search support for messages in a chat room makes it hard to get back to important messages. Ideally, we need a `#hash` tag based search. For messages with insufficient tagging, a smart tag extraction would enrich search.

The `indexer` plugin stores the message indexed by user defined `#hash` tags. If the tags are fewer than 5% of the words in the message, we enrich it using `prose` library based on extracted keywords with highest frequency.

`indexer` allows a mechanism to ignore indexing messages with `#hash` tags by specifying any of `@search`, `@ignore`, `@silent` or `@quiet`


#### Search `/search`

Simple tag based search for indexed data stored by `indexer` plugin.
The current state of implementation does `OR` and `AND` based search. The results are scoped with in a room.

Using `+` in the search query will initiate `AND` based query where all tags are matched.

By default, the search is performed across all messages in a chat room. Providing `scope=self` in the query URL limits the search scope to messages sent by the searcher. This can be used to implement `starred` messages.

The search results will be sent to a dedicated room registered by the user in the absence of an explicit `target` option in the query URL

### Artifacts and deployment

It is currently deployed in [`Atlassian dev-west2`](https://psyche.us-west-2.dev.atl-paas.net
) environment
