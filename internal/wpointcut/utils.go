package wpointcut

// joinPointMapMergeCallback is an auxiliary type that defines a filter join-point callback.
type joinPointMapMergeCallback func(context *PointcutContext) *PointcutContext

// joinPointMergeAux is an auxiliary function to merge join-points.
type joinPointMergeAux struct {
	context *PointcutContext
	left    joinPointMapMergeCallback
	right   joinPointMapMergeCallback
}

// newJoinPointMapMerge is a constructor for joinPointMergeAux.
func newJoinPointMapMerge(context *PointcutContext, left, right joinPointMapMergeCallback) *joinPointMergeAux {
	return &joinPointMergeAux{
		context: context,
		left:    left,
		right:   right,
	}
}

// and executes the merge as the logical operator AND.
func (merge *joinPointMergeAux) and() *PointcutContext {
	l := merge.left(merge.context)
	r := merge.right(l)
	return r
}

// or executes the merge as the logical operator OR.
func (merge *joinPointMergeAux) or() *PointcutContext {
	ctx := merge.context.clone()
	return merge.left(ctx).append(merge.right(ctx))
}
