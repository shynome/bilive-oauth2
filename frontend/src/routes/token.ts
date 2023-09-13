const TokenStoreKey = 'oauth-vierified-token'

export const save = (token: string) => {
	localStorage.setItem(TokenStoreKey, token)
}

export const get = () => {
	return localStorage.getItem(TokenStoreKey)
}

export const clear = () => {
	localStorage.removeItem(TokenStoreKey)
}
