# TokenCC

A typical workflow, starting with the sample token specification file, is shown in the following steps.
1.	Init the token chaincode, pass the user org_id and user_id of token admin. The org_id value is the organization ID or the membership service provider (MSP) ID of the token user. The user_id value is the user name or email address of the token user.
2.	Set up accounts and roles to use to test the token chaincode.
3.	Use the token admin credential.
a.	Call the InitializeFiatMoneyTOKToken method to initialize a new token.
b.	Use the CreateAccount method to create two accounts for token transactions. When you create an account, you specify the token_id, org_id, and user_id values. 
b.	Call the AddRole method to assign the minter role to the first user. In this scenario, only the first user is able to mint tokens.
3.	Test the token chaincode by minting and transferring tokens.
a.	Log in as the user with the minter role (the first user).
b.	Call the IssueTokens method to mint tokens to the first user's account. The amount of tokens that you can mint must be less than or equal to the max_mint_quantity value that is specified in the FiatMoneyToken.yml file.
c.	Call the TransferTokens method to transfer tokens from the first user to the second user.
d.	Call the GetAccountBalance method and check the result to see the new balance of the first user's account.
e.	Call the GetAccountTransactionHistory method and check the result to see the transaction history of the first user's account.
4.	Test the token chaincode by burning tokens.
a.	Log in as the second user, who has no roles assigned.
b.	Call the BurnTokens method to burn tokens in the second user's account. Because the burner role is not specified in the specification file, both the first user and the second user can burn tokens in their own accounts.
c.	Call the GetAccountBalance method and check the result to see the new balance of the second user's account.
d.	Call the GetAccountTransactionHistory method and check the result to see the transaction history of the second user's account.

