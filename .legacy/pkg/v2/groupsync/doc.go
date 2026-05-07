// This package provides utility to deal with the group synchronization issue.
//
// The "groups" attribute of the User resource is a readOnly attribute, which shall be updated according to the change
// of "members" in Group resources. This package provides mere utilities that may be helpful, it does not assume a
// certain way to resolve this issue.
package groupsync
