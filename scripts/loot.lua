-- loot database
--
-- loot.make:
-- Load a dagger, itemid("1") on a map @ pos 194, 254:
--
--     items.load({loot.make("1", 194, 254)})
--

-- Add something to the package path to include the itemdb.
package.path = package.path .. ";../game/data/?.lua"

function file_load(location, filename)
    local path = location .. "/" .. filename
    local f = assert(io.open(path, "r"))
    local c = f:read "*a"

    f:close()

    return c
end

itemdb = require('itemdb')
local DB = itemdb.DB

-- creates an item entry to place itemid on the map at (posx, posy)
local make = function(itemid, posx, posy)
    local index = itemid + 1
    i = DB[index]
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
