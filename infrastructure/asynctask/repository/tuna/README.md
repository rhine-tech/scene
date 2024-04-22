# Tuna

**An implementation of TaskDispatcher**

The tuna package is designed to efficiently manage the execution of asynchronous tasks. 

It consists of a task dispatcher and a Thunnus type that manages concurrent workers for executing tasks.

## Features
- Concurrent execution of tasks.
- Dynamic resizing of worker pool.
- Easily track the status of tasks.
- Reusable task definitions.
- Utilizes a separate asynctask package for defining task functions and statuses.

## Usage

### Creating a Thunnus

You can create a new `Thunnus` instance with a specific number of concurrent workers.

```go
thunnus := tuna.NewThunnus(16)
```

### Running Tasks

To run tasks, you can either use `Run` or `RunTask`. Both functions take a function conforming to the `asynctask.TaskFunc` type signature.

```go
thunnus.Run(func() error {
    // Your code here
    return nil
})
```

### Stopping and Resizing

You can stop the `Thunnus` or resize the worker pool.

```go
thunnus.Stop()
thunnus.Resize(10)
```

### Task Status

Tasks use the `asynctask.Task` type which includes functions to get and set the status.

```go
task := thunnus.Run(func() error {
    return nil
})
status := task.Status()
```

## License

MIT

## Author

@Aynakeya