#!/bin/sh

export scriptpath="../../scripts/?.lua"
export listener=":61507"
export factotum_driver="rpc"
export factotum_spec=""
export db_driver="sqlite3"
export db_spec="goland.db"
#export debug="true"
exec ./server -v=2 -logtostderr=true

