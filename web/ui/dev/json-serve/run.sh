#!/bin/bash

npx json-server db_presentation_server.json -p 3010 -r ps_routes.json -m middlewares.js
