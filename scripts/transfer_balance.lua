-- KEYS[1] = balance:<source_user_id>
-- KEYS[2] = version:<source_user_id>
-- KEYS[3] = balance:<dest_user_id>
-- KEYS[4] = version:<dest_user_id>
-- ARGV[1] = amount
-- ARGV[2] = expected source version
-- ARGV[3] = expected dest version

local source_balance = tonumber(redis.call("GET", KEYS[1]))
local source_version = tonumber(redis.call("GET", KEYS[2]))
local dest_balance = tonumber(redis.call("GET", KEYS[3]))
local dest_version = tonumber(redis.call("GET", KEYS[4]))
local amount = tonumber(ARGV[1])

if source_balance == nil or source_version == nil or dest_balance == nil or dest_version == nil then
  return {err = "MISSING_KEY"}
end

if source_balance < amount then
  return {err = "INSUFFICIENT_FUNDS"}
end

if source_version ~= tonumber(ARGV[2]) or dest_version ~= tonumber(ARGV[3]) then
  return {err = "VERSION_MISMATCH"}
end

redis.call("SET", KEYS[1], source_balance - amount)
redis.call("INCR", KEYS[2])

redis.call("SET", KEYS[3], dest_balance + amount)
redis.call("INCR", KEYS[4])

return {"OK", source_version + 1, dest_version + 1}
