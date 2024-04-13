const TokenStoreKey = 'oauth-vierified-token-v2'

export const save = (token: string) => {
	localStorage.setItem(TokenStoreKey, token)
}

export const get = () => {
	return localStorage.getItem(TokenStoreKey)
}

export const clear = () => {
	localStorage.removeItem(TokenStoreKey)
}
