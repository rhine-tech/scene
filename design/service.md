# Service Layer Design Principle

## Functionalities

- Service Layer is the layer that handle business logic.
- However, model related logic should be handled by the model itself. (in the domain layer)

## Error Handling

- Service Layer can only return errcode error, defined in its own domain.
- Service Layer **should** log all error from repository layer, but not compulsory.
- Service Layer **must** throw the error (if exist) to the delivery layer.