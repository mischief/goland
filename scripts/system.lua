-- things to NOT do
--
-- dont assign newindex to userdata
--gameserver.onpacket = 'lol'
--
-- use . not : for calling go userdata
--gameserver:ANYTHING()

package.path = package.path .. ";" .. gs.GetScriptPath() --";../?.lua"

-- item handling stuff?
items = require('items')
coll = require('collision')

collide = coll.collide

-- load our only map..
map = require('map1')

map.load()

-- get a debug shell after loading
debug.debug()

