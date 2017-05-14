# Logging
In core0 terminology, logging means capturing the running processes outputs and store or
forward it to loggers.

A logger can decide to print the output of the comman on terminal, or push it to redis.
The logger decided what messages to be logged based on the logger configurations, which can 
be overriden by the `command` `log_levels` attribute (if set) to force all loggers to caputre
and process these specific levels.

The default configuration template of `core0` will log all messages of levels `[1, 2, 4, 7, 8, 9]`
to both console and `redis`

CoreX logging is not configurable, it simply forwards all logs to core0 logging. Which means
Core0 logging configuration applies to both `core0` and `coreX` domains.

# Logging Messages
When running any process on core0/coreX the output of the process is captured
and processed as log messages. By default messages that are output on `stdout` stream
 are considered of level `1`, messages that are outputed on `stderr` stream are defaulted to
 level `2` messages.

The running process can leverage on the ability of core to process and handle different
log messages level by prefixing the output with the desired level as

```
8::Some message text goes here
```

Or for multiline output bulk
```
20:::
{
    "description": "A structured json output from the process",
    "data": {
        "key1": 100,
    }
}
:::
```

Using specific levels, u can pipe your messages through a different paths based on your nodes.

Also all `result` levels will make your return data catpured and set in the `data` attribute
 of your job result object. 

#Log Levels
- 1: stdout
- 2: stderr
- 3: message for endusers / public message
- 4: message for operator / internal message
- 5: log msg (unstructured = level5, cat=unknown)
- 6: log msg structured
- 7: warning message
- 8: ops error
- 9: critical error
- 10: statsd message(s)
- 20: result message, json
- 21: result message, yaml
- 22: result message, toml
- 23: result message, hrd
- 30: job, json (full result of a job)