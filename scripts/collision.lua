-- collision handling routines

local collidefns = {}

-- called when a flag hits something
collidefns.flag = function(o1,o2)

  return true
end

-- called when a block hits something
collidefns.block = function(o1,o2)
  gs.LuaLog("block collision: %s -> %s", o1.GetName(), o2.GetName())

  -- if a player hits a block...
  if o2.GetTag("player") == true then

    -- find out where we should move and if something is there
    x1, y1 = o1.GetPos()
    x2, y2 = o2.GetPos()

    if x1 ~= x2 or y1 ~= y2 then
      newx = x2 - (x1 - x2)
      newy = y2 - (y1 - y2)

      gs.LuaLog("%s move to %f %f", o1.GetName(), newx, newy)
      o1.SetPos(newx, newy)

    end


  end

  return true
end

-- called when a player hits something
collidefns.player = function(o1,o2)
  -- players can move blocks
  --
  if o2.GetName() == 'block' then
    x1,y1 = o1.GetPos()
    gs.LuaLog("o1 %.0f %.0f", x1, y1)
    x2,y2 = o2.GetPos()
    gs.LuaLog("o2 %.0f %.0f", x2, y2)

    if x1 ~= x2 or y1 ~= y2 then
      newx = x2-(x1-x2)
      newy = y2-(y1-y2)
      gs.LuaLog("%s move to %f %f", o2.GetName(), newx, newy)
      o2.SetPos(newx, newy)
    end
  elseif o2.GetName() == 'flag' then
    gs.LuaLog("%s found a flag!", o1.GetName())
  end

  return true
end

-- find the appropriate collision handler
local findcollisionfn = function(o1) 
  fn = collidefns[o1.GetName()]
  if not fn and o1.GetTag("player") == true then
    fn = collidefns['player']
  end
  return fn
end

-- collision entrypoint from go
-- o1 - the object that was moving
-- o2 - the object that was hit by o1
local collide = function(o1, o2)
  gs.LuaLog("colliding %s and %s", o1.GetName(), o2.GetName())

  -- get collision functions for both objects and call them
  fn1 = findcollisionfn(o1)
  fn2 = findcollisionfn(o2)
  if fn1 ~= nil then
    gs.LuaLog("collision: %s -> %s", o1.GetName(), o2.GetName())
    res1 = fn1(o1, o2)
  end
  if fn2 ~= nil then
    gs.LuaLog("collision: %s -> %s", o2.GetName(), o1.GetName())
    res2 = fn2(o2, o1)
  end

  return (fn1 and res1 or true) and (fn2 and res2 or true)

end

return {
  collide = collide,
  findfn = findcollisionfn,
  fns = collidefns,
}
