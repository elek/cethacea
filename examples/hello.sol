pragma solidity ^0.8.17;

contract hello {
   constructor() {
   }


   function one(bool t) public pure returns (string memory) {
       if (t) {
           return "true";
       } else {
           return "false";
       }
   }

}
