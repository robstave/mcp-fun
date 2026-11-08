package handlers

// Instructions is the short usage guide sent to clients during initialization.
const Instructions = `This server provides arithmetic, greetings, mock weather, and shared-deck card tools.
get_weather always returns mock data (76 degrees), never a live forecast.
draw_card and reshuffle_deck mutate one shared in-memory deck for this server process; restarting the server loses its saved deck ID. draw_card defaults to one card and requires a positive integer count. reshuffle_deck restores drawn cards.
Check MCP tool errors and the card API's success field. success:false is returned as normal content. Use only returned cards; do not invent missing results or automatically retry an uncertain draw, since the remote deck may already have changed.
Individual card operations are serialized, but separate tool calls are not an atomic round. Draw sequentially when assigning hands to players, and do not reshuffle between hands.
Use the poker_night prompt for a five-card showdown among Silas, Clementine, Horace, and Beatrice. Retrieving the prompt only returns instructions; the chat must call the tools and grade the hands using standard poker rules, including kickers and ties.`
