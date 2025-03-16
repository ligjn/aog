===============================
AOG Key Mechanisms
===============================


In this section, we will introduce two important mechanisms of ``AOG`` for two
critical challenges we mentioned earlier: :ref:`compatibility_issue` and
:ref:`availability_issue`.

In addition, we will discuss about how to match the models.


.. _flavor_conversion:

Conversion of API Flavors for Compatibility
===============================================================

``AOG`` will convert the requests and responses if the ``API Flavor`` of
Application can be different from the ``API Flavor`` of the underlying ``Service
Provider``. 

There are several possible scenarios as listed below. In each scenario,
different conversion will be done by ``AOG``. The most complicated one is
scenario C, where neither the application nor the service provider uses the
``AOG Flavor``. In this case, ``AOG`` will convert the request from the app's
flavor to the ``AOG Flavor`` first, then convert it to the service provider's
flavor. The response will be converted in the reverse order.


.. list-table:: Conversion of API Flavors for Compatibility
   :header-rows: 1
   :widths: 10 10 10 100

   * - Situation
     - App's Flavor
     - Service Provider's Flavor
     - Conversion done by AOG
   * - A
     - X
     - X
     - None
   * - B
     - AOG
     - AOG
     - None
   * - C
     - X
     - Y
     - | Request: X -> AOG then AOG -> Y
       | Response: Y -> AOG then AOG -> X
   * - D
     - AOG
     - Y
     - | Request: AOG -> Y
       | Response: Y -> AOG
   * - E
     - X
     - AOG
     - | Request: X -> AOG
       | Response: AOG -> X




.. graphviz::
    :align: center
    
    digraph G {
        rankdir=TB
        compound=true
        label = "Situations of API Flavors"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 10 ]


        subgraph cluster_a {
            label = "A"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_a[label="App", fillcolor="#eeeeff"]
            aog_a[label="AOG API Layer", fillcolor="#ffffcc"]
            sp_a[label="Service Provider", fillcolor="#ffcccc"]

            app_a -> aog_a [label=" X", dir=both]
            aog_a -> sp_a [label=" X", dir=both]

        }


        subgraph cluster_b {
            label = "B"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_b[label="App", fillcolor="#eeeeff"]
            aog_b[label="AOG API Layer", fillcolor="#ffffcc"]
            sp_b[label="Service Provider", fillcolor="#ffcccc"]

            app_b -> aog_b [label=" AOG", dir=both]
            aog_b -> sp_b [label=" AOG", dir=both]
        }

        subgraph cluster_c {
            label = "C"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_c[label="App", fillcolor="#eeeeff"]
            aog_c[label="AOG API Layer", fillcolor="#ffffcc"]
            sp_c[label="Service Provider", fillcolor="#ffcccc"]

            app_c -> aog_c [label=" X", dir=both]
            aog_c -> sp_c [label=" Y", dir=both]

        }


        subgraph cluster_d {
            label = "D"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_d[label="App", fillcolor="#eeeeff"]
            aog_d[label="AOG API Layer", fillcolor="#ffffcc"]
            sp_d[label="Service Provider", fillcolor="#ffcccc"]

            app_d -> aog_d [label=" AOG", dir=both]
            aog_d -> sp_d [label=" Y", dir=both]
        }

        subgraph cluster_e {
            label = "E"
            color="#dddddd"
            fillcolor="#eeeeee"

            app_e[label="App", fillcolor="#eeeeff"]
            aog_e[label="AOG API Layer", fillcolor="#ffffcc"]
            sp_e[label="Service Provider", fillcolor="#ffcccc"]

            app_e -> aog_e [label=" X", dir=both]
            aog_e -> sp_e [label=" AOG", dir=both]
        }

    }



A more detailed flow is illustrated here, for conversion of requests and
responses respectively.


.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Request Body in AOG API Layer"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, color="#ffffcc", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

        receive [label="AOG \nReceives \nApp's \nRequest"]
        is_same_flavor [label="App's Flavor\n==\nFlavor of \nService \nProvider ?", shape=diamond]
        is_app_aog [label="App's Flavor \n==\nAOG ?", shape=diamond]
        is_sp_aog [label="Flavor of \nService\nProvider\n==\nAOG ?", shape=diamond]
        to_aog [label="convert\nRequest\nto\nAOG\nFlavor"]
        from_aog [label="convert to\nFlavor of\nService\nProvider"]
        invoke [label="Invoke\nService\nProvider\nwith its\nFlavor"]

        receive->is_same_flavor
        is_same_flavor->invoke [label="Yes"]
        is_same_flavor->is_app_aog [label="No"]
        is_app_aog -> is_sp_aog [label="Yes"]
        is_app_aog -> to_aog [label="No"]
        to_aog -> is_sp_aog
        is_sp_aog -> invoke [label="Yes"]
        is_sp_aog -> from_aog [label="No"]
        from_aog -> invoke

        subgraph r1 {
            rank="same"
            receive, is_same_flavor, invoke
        }

        subgraph r2 {
            rank="same"
            is_app_aog, to_aog, is_sp_aog, from_aog
        }
    }



.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Response Body in AOG API Layer"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, color="#ffffcc", style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

        receive [label="AOG \nReceives \nResponse\nfrom\nService\nProvider"]
        is_same_flavor [label="App's Flavor\n==\nFlavor of \nService \nProvider ?", shape=diamond]
        is_app_aog [label="App's Flavor\n==\nAOG ?", shape=diamond]
        is_sp_aog [label="Flavor of \nService\nProvider\n==\nAOG ?", shape=diamond]
        to_aog [label="convert\nResponse\nto\nAOG\nFlavor"]
        from_aog [label="convert\nto\nApp's\nFlavor"]
        send [label="Send\nResponse\nin App's\nFlavor\nto App"]

        receive->is_same_flavor
        is_same_flavor->send [label="Yes"]
        is_same_flavor->is_sp_aog [label="No"]
        is_sp_aog -> is_app_aog [label="Yes"]
        is_sp_aog -> to_aog [label="No"]
        to_aog -> is_app_aog
        is_app_aog -> send [label="Yes"]
        is_app_aog -> from_aog [label="No"]
        from_aog -> send 

        subgraph r1 {
            rank="same"
            receive, is_same_flavor, send
        }

        subgraph r2 {
            rank="same"
            is_app_aog, to_aog, is_sp_aog, from_aog
        }
    }



.. graphviz:: 
    :align: center

    digraph G {
        rankdir=TB
        compound=true
        label = "Conversion of Request Body in Service Provider"
        graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
        node [fontname = "Verdana", fontsize = 12, shape=box, style="filled", penwidth=0.5]
        edge [fontname = "Verdana", fontsize = 12 ]

    }


.. _hybrid_scheduling:

Hybrid Scheduling for Availability
========================================================

``AOG`` provides hybrid scheduling, i.e. when needed, it will dispatch
application's request (with necessary conversion) to a remote alternative ``AOG
Service Provider`` (usually a cloud service) instead of local. This is very
helpful when local AIPC is busy, or the desired service is not provided by
current PC, or the user wants to use VIP service at cloud etc.

``AOG`` makes such dispatch decision by following the specified ``hybrid
policy``. The AIPC with ``AOG`` installed has a system-wide configuration (See
:doc:`/aog_platform_config`) which specifies all of the available ``AOG
Service`` and their corresponding local and remote ``AOG Service Providers``,
along with the default ``hybrid policy`` to switch between these providers. 

Furthermore, the application can also overwrite the default ``hybrid policy``
defined by the platform config. For example, the application may force to use
the cloud service for a particular request, it can then add ``hybrid_policy:
always_remote`` in the JSON body of request to send.



.. graphviz:: 
   :align: center

   digraph G {
     rankdir=TB
     compound=true
     label = "Hybrid Scheduling"
     graph [fontname = "Verdana", fontsize = 10, style="filled", penwidth=0.5]
     node [fontname = "Verdana", fontsize = 10, shape=box, color="#333333", style="filled", penwidth=0.5] 

     app[label="Application", fillcolor="#eeeeff"]
     aog[label="AOG to Dispatch - based on Hybrid Policy", fillcolor="#ffffcc"]
     local[label="Local AOG Service Provider", fillcolor="#ffcccc"]
     cloud[label="Remote AOG Service Provider", fillcolor="#ffcccc"]

     app -> aog

     aog -> local[style="dashed"]
     aog -> cloud[style="dashed"]

   }




.. _match_models:

Match Models
========================================================

In a lot of situations, the application may want to specify the preferred model
to use, but the underlying ``AOG Service Provider`` either doesn't provide the 
model, or it provides the model but the name is slightly different.

Currently ``AOG`` provides a simple mechanism which tries to pick the model from 
the service provider which best matches the required model by application. This 
is up to change or evolve in the future.

First, when defines the available ``AOG Service Provider``, the
:doc:`/aog_platform_config` can also list the available models for each service
provider, as part of its :ref:`Property of AOG Service Provider
<aog_service_provider_properties>`.

Then, the application can specify the model name in the request, for example,
``model: xx-7B`` in its JSON body of the request. ``AOG`` will do a fuzz match
between this expected model and the available models of the service provider,
and ask to use the most similar one. 

