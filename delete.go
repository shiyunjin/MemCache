package memory

func (c *Cache) Delete(key string) error {
	c.storage.Del([]byte(key))
	return nil
}
