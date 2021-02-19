package bspwmnode

// TODO
//  But I'll likely need this code:

/*
type resResolverMatcher struct {
	t   *testing.T
	res interface{}
}

// TODO: Add the below to a `bspctest` package inside the `bspc-go` repo, akin to `zap`'s `zaptest`.
func ResponseResolver(t *testing.T, res interface{}) *resResolverMatcher {
	return &resResolverMatcher{
		t:   t,
		res: res,
	}
}

func (m *resResolverMatcher) String() string {
	bb, err := json.Marshal(m.res)
	require.NoError(m.t, err)

	return string(bb)
}

func (m *resResolverMatcher) Matches(x interface{}) bool {
	resolver, ok := x.(bspc.QueryResponseResolver)
	if !ok {
		return false
	}

	bb, err := json.Marshal(m.res)
	require.NoError(m.t, err)

	// Running this populates the variable passed into it by reference.
	err = resolver(bb)
	require.NoError(m.t, err)

	return true
}
*/
