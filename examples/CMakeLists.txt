include_directories(./)

add_executable(apixsrv apixsrv.c opt.c)
target_link_libraries(apixsrv apix)

add_executable(apixcli apixcli.c opt.c cli.c)
target_link_libraries(apixcli apix readline pthread)

add_executable(echo-server echo-server.c)
target_link_libraries(echo-server apix)

install(TARGETS apixsrv apixcli
    RUNTIME DESTINATION bin
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    PUBLIC_HEADER DESTINATION include
)
