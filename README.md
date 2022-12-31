# Proxy

This proxy server is a program that allows you to forward HTTP requests to a target address and modify the responses received from the target.

It has the following features:

1. The ability to specify the target and port for the proxy using command-line flags.
2. The option to add basic authentication to the proxy using command-line flags.
3. The ability to enable verbose output using a command-line flag. This will display messages in the command prompt showing the actions of the proxy.
4. The option to decrypt responses from the target using an encryption key provided as a hexadecimal string.
5. The ability to cache decrypted responses in memory using the groupcache package. The cache size and eviction policy can be customized.

## Instructions

###### Building

To build the program, use the following command:

```
go build proxy.go

```

###### Usage

```
Syntax: proxy.exe [-target TARGET] [-port PORT] 
                  [-username USERNAME] [-password PASSWORD] [-key KEY] [-verbose]

[*] Options:
  Specify the target address
    -target TARGET (required)
    
  Specify the port used
    -port PORT (optional, default port: 8080)       
    
  Specify the username if authentication is needed  
    -username USERNAME (optional)              
    
  Specify the password if authentication is needed  
    -password PASSWORD (optional)           
  
  Specify the AES encrypion key. The key should be a hexadecimal string.
    -key KEY (optional)         
    
  Enables verbose output  
    -verbose (optional)                             

[*] Examples:
  Specify the target to "example.com" and port to "8081" for the proxy
    proxy.exe -target=example.com -port=8081        
    
  Specify the basic authentication username to "user" and password to "pass" 
  for the proxy (Note! Target is missing from this example but is required)  
    proxy.exe -username=user -password=pass         
    
  Enables verbose output (Note! Target is missing from this example but is required)  
    proxy.exe -verbose                          
    
  Specifies the AES encrypion key used for decryption (Note! Target is missing from this example but is required)  
    proxy.exe -key=6368616e676520746869732070617373776f726420746f206120736563726574
                                                  
```

The proxy server will cache decrypted responses in memory using the groupcache package. 
The cache size and eviction policy can be customized by modifying the lruCache variable at the beginning of the main function.
