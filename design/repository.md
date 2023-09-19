# Repository Layer Design Principle

## Functionalities

- Repository Layer is the layer that handle all external interaction.

## Error Handling
- Repository Layer can return any error you want. It can be errcode error or random error.
- Errors can be logged in the repository layer. But must throw the error to the service layer.