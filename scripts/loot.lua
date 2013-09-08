-- loot database
--
-- loot.make:
-- Load a dagger, itemid(1) on a map @ pos 194, 254:
--
--     items.load({loot.make(1, 194, 254)})
--

function file_load(location, filename)
    local path = location .. "/" .. filename
    local f = assert(io.open(path, "r"))
    local c = f:read "*a"

    f:close()

    return c
end

local db_string = file_load('../game/data', 'loot.json')
local DB = Json.Decode(db_string)

local get_item = function(DB, itemid)
    return DB.items[itemid + 1]
end
-- creates an item entry to place itemid on the map at (posx, posy)
local make = function(itemid, posx, posy)
    i = get_item(DB, itemid)
    return { i.name, posx, posy, i.glyph, i.color_fg, i.color_bg }
end

-- modifies the color of the item (useful for flags)
local colorize = function(item, color)
    item[5] = color
    return item
end

return {
    make = make,
    colorize = colorize
}
