-- map1.lua - first map yay

collision = require('collision')

-- the goods
flags = {
  {'flag', 119, 125, '†', 'red', ''},
  {'flag', 119, 133, '†', 'red', ''},
  {'flag', 129, 125, '†', 'blue', ''},
  {'flag', 129, 133, '†', 'blue', ''},
}

blocks = {
  {'block', 122, 130, '¤', 'red', ''},
  {'block', 124, 128, '¤', 'red', ''},
  {'block', 124, 132, '¤', 'blue', ''},
  {'block', 126, 130, '¤', 'blue', ''},
}

score = {
  {'scorepoint', 97, 115, '_', 'white', 'red'},
  {'scorepoint', 124, 130, '_', 'white', 'green'},
  {'scorepoint', 152, 140, '_', 'white', 'blue'},
}

scores = {}

collision.fns.scorepoint = function(o1, o2)
  if o2.GetTag('player') ~= true then
    return false
  end

  gs.LuaLog("%s has stepped on a %s", o2.GetName(), o1.GetName())

  subobjs = o2.SubObjects.GetSlice()

  print(type(subobjs))
  slice = subobjs
  --gs.LuaLog("2 %s %f", type(slice), #slice)
  --slice = subobjs.GetSlice()
  print(type(slice), #slice)
  for k,v in pairs(getmetatable(slice)) do
    print(k,v)
  end

  -- does not work :(
  --for k,v in ipairs(slice) do
  --  print(k,v)
  --end
  --obj = slice[0]
  --print(type(obj))
  --slice = assert(slice:Slice(0, #slice)) -- XXX why does this work????
--  print('slice has ', #slice)

  for i=2, #slice do
    obj = slice[i]
--    print(type(obj))

    -- if the player has a flag, take it and give them a point
    -- then spawn another flag
    if obj.GetName() == 'flag' then

      gs.LuaLog("%s has flag, removing", o2.GetName())
      o2.RemoveSubObject(obj)

      -- move the flag to random point

      newx, newy = math.random(119, 129), math.random(125,133)
      obj.SetPos(newx, newy)
      obj.SetTag("visible", true)
      obj.SetTag("gettable", true)

      -- tell everyone the object changed
      gs.SendPkStrAll("Raction", obj)

      -- give player a point
      pname = o2.GetName()
      if scores[pname] ~= nil then
        scores[pname] = scores[pname] + 1
      else
        scores[pname] = 1
      end

      gs.SendPkStrAll("Rchat", string.format("ctf: %s has %d points!", pname, scores[pname]))


    end
  end

  return true
end

fns = {}

fns.load = function()
  gs.LoadMap('../server/map')
  items.load(flags)
  items.load(blocks)

  -- load the score point
  for _, v in pairs(score) do
    sp = object.New(v[1])
    sp.SetPos(v[2], v[3])
    sp.SetTag('visible', true)
    sp.SetGlyph(util.NewGlyph(v[4], v[5], v[6]))
    gs.AddObject(sp)
  end
end

return fns

