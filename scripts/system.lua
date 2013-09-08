-- things to NOT do
--
-- dont assign newindex to userdata
--gameserver.onpacket = 'lol'
--
-- use . not : for calling go userdata
--gameserver:ANYTHING()

-- item handling stuff?
--JSON = require('JSON')
Json = require('json')
loot = require('loot')
items = require('items')
coll = require('collision')

collide = coll.collide

-- load our only map..
map = require('map1')

map.load()

-- get a debug shell after loading
if gs.Debug() == true then
  debug.debug()
end

