# Job Commands

Available commands:

- [job.list](#list)
- [job.kill](#kill)
- [job.unschedule](#unschedule)

<a id="list"></a>
## job.list

Lists all running jobs.

Arguments:
```javascript
{
  'id': {id},
}
```

Values:
- **id**: Optional parameter in order to list only one specific job

<a id="kill"></a>
## job.kill

Signals a job with given ID, and signal number

Arguments:
```javascript
{
  'id': {id},
  'signal': {signal},
}
```

<a id="unschedule"></a>
## job.unschedule
If you started a job with `recurring_period` set, unschedule will prevent it from restarting 
once it dies. It does not kill the running job, just mark it to not restart again once it exits.

Usually u will follow a call to unschedule to a call to kill to stop the process completely.
