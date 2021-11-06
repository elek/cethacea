pragma solidity ^0.8.7;

import "./openzeppelin/contracts/token/ERC20/ERC20.sol";

contract GLDToken is ERC20 {
    constructor() ERC20("Gold", "GLD") public {
        _mint(msg.sender, 1000);
    }
}
