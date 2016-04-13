import Api from '../Api';
import {
    REQUEST_NODE_REVISIONS,
    RECEIVE_NODE_REVISIONS,
    INVALIDATE_NODE_REVISIONS,
    SELECT_NODE_REVISION,
    REQUEST_NODE_REVISION,
    RECEIVE_NODE_REVISION
} from '../constants/ActionTypes';


// —————————————————————————————————————————————————————————————————————————————————————————————————————————————————————
// Revisions
// —————————————————————————————————————————————————————————————————————————————————————————————————————————————————————
function requestNodeRevisions(uuid, page) {
    return {
        type: REQUEST_NODE_REVISIONS,
        uuid,
        page
    };
}

function receiveNodeRevisions(uuid, {
    elements,
    page,
    next
}) {
    return {
        uuid,
        type:     RECEIVE_NODE_REVISIONS,
        items:    elements,
        page,
        nextPage: next
    };
}

function fetchNodeRevisions(uuid, page) {
    return (dispatch, getState) => {
        dispatch(requestNodeRevisions(uuid, page));
        Api.nodeRevisions(uuid, page, getState().security.token)
            .then(response => {
                dispatch(receiveNodeRevisions(uuid, response));
            })
        ;
    };
}

function shouldFetchNodeRevisions(state, uuid, page) {
    const revisions = state.nodesRevisionsByUuid[uuid];
    if (!revisions) {
        return true;
    }

    if (revisions.isFetching) {
        return false;
    }

    if (revisions.page < page) {
        return true;
    }

    return revisions.didInvalidate;
}

export function fetchNodeRevisionsIfNeeded(uuid, page = 1) {
    return (dispatch, getState) => {
        if (shouldFetchNodeRevisions(getState(), uuid, page)) {
            dispatch(fetchNodeRevisions(uuid, page));
        }
    };
}

export function invalidateNodeRevisions(uuid) {
    return {
        type: INVALIDATE_NODE_REVISIONS,
        uuid
    };
}


// —————————————————————————————————————————————————————————————————————————————————————————————————————————————————————
// Revision
// —————————————————————————————————————————————————————————————————————————————————————————————————————————————————————
export function selectNodeRevision(uuid, id) {
    return {
        type: SELECT_NODE_REVISION,
        uuid,
        id
    };
}

function requestNodeRevision(uuid, id) {
    return {
        type: REQUEST_NODE_REVISION,
        uuid,
        id
    };
}

function receiveNodeRevision(revision) {
    return {
        type: RECEIVE_NODE_REVISION,
        revision
    };
}

function fetchNodeRevision(uuid, id) {
    return (dispatch, getState) => {
        dispatch(requestNodeRevision(uuid, id));
        Api.nodeRevision(uuid, id, getState().security.token)
            .then(revision => {
                dispatch(receiveNodeRevision(revision));
            })
        ;
    };
}

function shouldFetchNodeRevision(state, uuid, id) {
    const revisions = state.nodesRevisionsByUuid[uuid];
    if (!revisions) {
        return true;
    }

    const revision = revisions.byRevisionId[id];
    if (!revision) {
        return true;
    }

    return revision.didInvalidate;
}

export function fetchNodeRevisionIfNeeded(uuid, id) {
    return (dispatch, getState) => {
        if (shouldFetchNodeRevision(getState(), uuid, id)) {
            dispatch(fetchNodeRevision(uuid, id));
        }
    };
}
