package retail

// PermRoot allows all actions on a tenant, only tenant owner account can have this permission,
// it's meant to prevent reach a state where none of the users have the necessary permissions
// to perform an operation.
const PermRoot = "tenant-root"
