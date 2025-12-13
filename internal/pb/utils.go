package pb

// GetMediaKey safely accesses the first mediaKey from hash lookup response
func (x *FindMediaByHashResponse) GetMediaKey() string {
	if x == nil || x.Field1 == nil || x.Field1.Field2 == nil || x.Field1.Field2.Field2 == nil {
		return ""
	}
	return x.Field1.Field2.Field2.MediaKey
}
