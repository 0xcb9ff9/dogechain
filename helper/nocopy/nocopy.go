// from https://github.com/golang/go/issues/8005#issuecomment-190753527

package nocopy

type NoCopy struct{}

func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}
