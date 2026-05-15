# anticheat

## Flow

unsure when but we will get a token from auth service -> game launches (unsure if bootstrapper yet or not) with a -a=token argument -> decrypt token & parse -> connect to
/socket/v1/anticheat/:accountId -> client will send Registered message (ofc heavy encryption) -> game engine init -> client will send a challenge -> server returns if it succeeded or not (if success -> proceed | if fail terminateprocess(getcurrentprocess(), 0)) -> if proceeds start a heartbeat
